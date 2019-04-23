package deployment

import (
	"context"
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log = logf.Log.WithName("controller_deployment")
)

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

	err = c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForObject{},
		&CanaryObjectPredicate{},
	)
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

func fallbackReconcile(err error) (reconcile.Result, error) {
	// we only want to requeue these errors
	if strings.Contains(err.Error(), "the object has been modified") {
		err = nil
	}

	return reconcile.Result{
		RequeueAfter: time.Second * 3,
		Requeue:      true,
	}, err
}

// Reconcile reads that state of the cluster for a Deployment object and makes changes based on the state read
// and what is in the Deployment.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	log := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	canaryInstance, err := r.reconcileCanaryDeployment(ctx, request.NamespacedName)
	if err != nil {
		// special watcher case, we don't want to retry deleted objects
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return fallbackReconcile(err)
	}

	canaryPolicy, err := r.GetCanaryPolicy(ctx, canaryInstance)
	if err != nil {
		return fallbackReconcile(err)
	}

	log.Info("Reconciling Deployment with policy: ", "policy-name", canaryPolicy.Name)

	// reconcile the canary configuration (2)
	_, err = r.reconcileCanaryConfig(ctx, canaryInstance, canaryPolicy)
	if err != nil {
		return fallbackReconcile(err)
	}

	// reconcile the baseline (3)
	_, err = r.reconcileBaseline(ctx, canaryInstance, canaryPolicy)
	if err != nil {
		return fallbackReconcile(err)
	}

	return reconcile.Result{}, nil
}

// GetCanaryPolicyName returns the canary policy name if it exists
func GetCanaryPolicyName(d *appsv1.Deployment) string {
	// get the canary configuration
	return d.ObjectMeta.Labels[CanaryPolicyLabel]
}

// GetCanaryConfigName returns the canary config name if it exists
func GetCanaryConfigName(d *appsv1.Deployment) string {
	return d.ObjectMeta.Labels[CanaryConfigLabel]

}

// MakeCanaryConfigName returns the canary config name based on the
// name of the canary instance and the policy
func MakeCanaryConfigName(p *v1.CanaryPolicy, c *appsv1.Deployment) string {
	return p.Name + "-" + c.Name
}

// MakeBaselinename returns the name of the baseline object
func MakeBaselineName(c *appsv1.Deployment) string {
	return c.Name + "-baseline"
}

func (r *ReconcileDeployment) UpdateCanaryConfigName(ctx context.Context, policy *v1.CanaryPolicy, c *appsv1.Deployment) error {
	c.ObjectMeta.Labels[CanaryConfigLabel] = MakeCanaryConfigName(policy, c)

	return r.client.Update(ctx, c)
}

func (r *ReconcileDeployment) reconcileCanaryDeployment(ctx context.Context, nn types.NamespacedName) (*appsv1.Deployment, error) {
	// Fetch the Deployment d
	d := &appsv1.Deployment{}
	err := r.client.Get(ctx, nn, d)
	if err != nil {
		return nil, err
	}

	canaryPolicy, err := r.GetCanaryPolicy(ctx, d)
	if err != nil {
		return nil, err
	}

	canaryConfigName := GetCanaryConfigName(d)
	// canary config not found, we should create it (1)
	if canaryConfigName == "" {
		// this will trigger a new reconcile due to the update
		err := r.UpdateCanaryConfigName(ctx, canaryPolicy, d)
		if err != nil {
			return nil, err
		}
	}

	return d, err
}

func (r *ReconcileDeployment) GetCanaryPolicy(ctx context.Context, c *appsv1.Deployment) (*v1.CanaryPolicy, error) {
	policyName := GetCanaryPolicyName(c)

	// get the policy
	canaryPolicy := &v1.CanaryPolicy{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: c.Namespace,
		Name:      policyName,
	}, canaryPolicy)

	// ignore it if the object was deleted
	if errors.IsNotFound(err) {
		return nil, nil
	}

	return canaryPolicy, err
}
