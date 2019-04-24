package canary

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	"github.com/petomalina/krane/pkg/apis/krane/v1"

	appsv1 "k8s.io/api/apps/v1"
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
	if strings.Contains(err.Error(), "the object has been modified") {
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

	instance := &v1.Canary{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ready, err := r.waitForCanaryAndBaseline(ctx, instance)
	if !ready {
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	_, err = r.reconcileDestinationRules(ctx, instance)
	if err != nil {
		return fallbackReconcile(err)
	}

	_, err = r.reconcileVirtualService(ctx, instance)
	if err != nil {
		return fallbackReconcile(err)
	}

	reqLogger.Info("Canary Config Reconciliation complete")

	return reconcile.Result{}, nil
}

func (r *ReconcileCanary) waitForCanaryAndBaseline(ctx context.Context, cfg *v1.Canary) (bool, error) {
	return true, nil
}

// IsDeploymentReady returns true if the deployment is available
func IsDeploymentReady(d *appsv1.Deployment) bool {
	if len(d.Status.Conditions) <= 0 {
		return false
	}

	return d.Status.Conditions[0].Type == "Available" && d.Status.Conditions[0].Status == "True"
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
