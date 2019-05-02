package canary

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	"github.com/petomalina/krane/pkg/apis/krane/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_canary")

// Add creates a new Canary Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCanary{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("canary-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Canary
	err = c.Watch(&source.Kind{Type: &v1.Canary{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		OwnerType:    &v1.Canary{},
		IsController: true,
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCanary{}

// ReconcileCanary reconciles a Canary object
type ReconcileCanary struct {
	client client.Client
	scheme *runtime.Scheme
}

func fallbackReconcile(err error) (reconcile.Result, error) {
	// we only want to requeue these errors
	if err != nil && strings.Contains(err.Error(), "the object has been modified") {
		err = nil
	}

	return reconcile.Result{
		RequeueAfter: time.Second * 5,
		Requeue:      true,
	}, err
}

// Reconcile reads that state of the cluster for a Canary object and makes changes based on the state read
// and what is in the Canary.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCanary) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Canary")

	canaryCfg := &v1.Canary{}
	err := r.client.Get(ctx, request.NamespacedName, canaryCfg)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	err = r.reconcileCanaryAndBaseline(ctx, canaryCfg)
	if err != nil {
		reqLogger.Info("Canary or Baseline deployments errored", "err", err.Error())
		return fallbackReconcile(err)
	}

	// wait for the deployments to be ready
	if canaryCfg.Status.Initialization.Status == v1.CanaryPhaseStatus_InProgress {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	_, err = r.reconcileTestJob(ctx, canaryCfg)
	if err != nil {
		reqLogger.Info("TestJob reconciliation error", "err", err.Error())
		return fallbackReconcile(err)
	}

	_, err = r.reconcileJudgeJob(ctx, canaryCfg)
	if err != nil {
		reqLogger.Info("Judge reconciliation error", "err", err.Error())
		return fallbackReconcile(err)
	}

	_, err = r.reconcileVirtualService(ctx, canaryCfg)
	if err != nil {
		reqLogger.Info("VirtualService reconciliation error", "err", err.Error())
		return fallbackReconcile(err)
	}

	err = r.ReconcilePhaseStatus(ctx, canaryCfg)
	if err != nil {
		return fallbackReconcile(err)
	}

	reqLogger.Info("Canary Config Reconciliation complete")

	return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *ReconcileCanary) reconcileCanaryAndBaseline(ctx context.Context, cfg *v1.Canary) error {
	canaryDeployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: cfg.Namespace,
		Name:      cfg.Spec.Deployments.Canary,
	}, canaryDeployment)
	if err != nil {
		return err
	}

	baselineDeployment := &appsv1.Deployment{}
	err = r.client.Get(ctx, types.NamespacedName{
		Namespace: cfg.Namespace,
		Name:      cfg.Spec.Deployments.Baseline,
	}, baselineDeployment)
	if err != nil {
		return err
	}

	// not available, don't start the test
	if canaryDeployment.Status.AvailableReplicas <= 0 || baselineDeployment.Status.AvailableReplicas <= 0 {
		cfg.Status.Initialization.Status = v1.CanaryPhaseStatus_InProgress
		cfg.Status.Initialization.Message = "Initializing baseline and canary"
	} else {
		cfg.Status.Initialization.Status = v1.CanaryPhaseStatus_Success
		cfg.Status.Initialization.Message = "Initialized baseline and canary deployments"
	}

	err = r.client.Status().Update(ctx, cfg)
	if err != nil {
		return err
	}

	return nil
}

// ReconcilePhaseStatus updates the phases of the canary
func (r *ReconcileCanary) ReconcilePhaseStatus(ctx context.Context, cfg *v1.Canary) error {
	L := log.WithValues("canary", cfg.Name)

	newStage := cfg.Status.Progress

	switch cfg.Status.Progress {
	case v1.CanaryProgress_Initializing:
		if cfg.Status.Initialization.Status == v1.CanaryPhaseStatus_Success {
			newStage = v1.CanaryProgress_Testing
			cfg.Status.Testing.Status = v1.CanaryPhaseStatus_Queued
		}
	case v1.CanaryProgress_Testing:
		if cfg.Status.Testing.Status == v1.CanaryPhaseStatus_Success {
			newStage = v1.CanaryProgress_Canary
			cfg.Status.Canary.Status = v1.CanaryPhaseStatus_Queued
		}
	case v1.CanaryProgress_Canary:
		if cfg.Status.Canary.Status == v1.CanaryPhaseStatus_Success {
			newStage = v1.CanaryProgress_Reporting
			cfg.Status.Reporting.Status = v1.CanaryPhaseStatus_Queued
		}
	case v1.CanaryProgress_Reporting:
		if cfg.Status.Reporting.Status == v1.CanaryPhaseStatus_Success {
			newStage = v1.CanaryProgress_Cleanup
			cfg.Status.Cleanup.Status = v1.CanaryPhaseStatus_Queued
		}
	}

	if cfg.Status.Progress != newStage {
		L.Info("Transitioning into new stage", "oldStage", cfg.Status.Progress, "newStage", newStage)
		cfg.Status.Progress = newStage
		return r.client.Status().Update(ctx, cfg)
	}

	// reset the progress if it gets lost (TODO)
	if cfg.Status.Progress == "" {
		L.Info("Updating initialization Progress")
		cfg.Status.Progress = v1.CanaryProgress_Initializing
		return r.client.Status().Update(ctx, cfg)
	}

	return nil
}

func (r *ReconcileCanary) GetCanaryPolicy(ctx context.Context, c *v1.Canary) (*v1.CanaryPolicy, error) {
	// get the policy
	canaryPolicy := &v1.CanaryPolicy{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      c.Spec.Policy,
	}, canaryPolicy)

	return canaryPolicy, err
}
