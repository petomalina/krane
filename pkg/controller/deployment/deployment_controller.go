package deployment

import (
	"context"
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	"k8s.io/apimachinery/pkg/types"

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

var log = logf.Log.WithName("controller_deployment")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Deployment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeployment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeployment{}

// ReconcileDeployment reconciles a Deployment object
type ReconcileDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Deployment object and makes changes based on the state read
// and what is in the Deployment.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Deployment")

	// Fetch the Deployment instance
	instance := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// get the canary configuration
	canaryPolicy, ok := instance.ObjectMeta.Labels["krane.sh/canary-policy"]
	if !ok {
		return reconcile.Result{}, nil
	}
	log.Info("Reconciling a Canary Deployment", "policy", canaryPolicy, "deployment", instance.Name)

	return reconcile.Result{}, nil
}

func (r *ReconcileDeployment) createBaselineDeployment(ctx context.Context, canary *appsv1.Deployment, policy *v1.CanaryPolicy) (*appsv1.Deployment, error) {
	var baseline *appsv1.Deployment

	// get the old deployment to retrieve containers
	base := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: canary.Namespace,
		Name:      policy.Spec.Base,
	}, base)
	if err != nil {
		return nil, err
	}

	// Default to new baseline mode
	if policy.Spec.BaselineMode == "" {
		policy.Spec.BaselineMode = v1.BaselineModeNew
	}

	switch policy.Spec.BaselineMode {
	case v1.BaselineModeNew:
		baseline = canary.DeepCopy()
		// copy over previous container configuration
		baseline.Spec.Template.Spec.Containers = base.Spec.Template.Spec.Containers

	case v1.BaselineModeOld:
		baseline = base.DeepCopy()
	}

	baseline.ObjectMeta.Name += "-baseline"
	// argh, golang, why you no support pointers
	singleReplica := int32(1)
	baseline.Spec.Replicas = &singleReplica
	// register release to be baseline
	baseline.ObjectMeta.Labels["release"] = "baseline"

	return baseline, nil
}
