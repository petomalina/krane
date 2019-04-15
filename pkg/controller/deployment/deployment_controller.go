package deployment

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ctx := context.Background()

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Deployment")

	// Fetch the Deployment canaryInstance
	canaryInstance := &appsv1.Deployment{}
	err := r.client.Get(ctx, request.NamespacedName, canaryInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// get the canary configuration
	canaryPolicyName, ok := canaryInstance.ObjectMeta.Labels["krane.sh/canary-policy"]
	if !ok {
		reqLogger.Info("Skipping reconcilement, no canary-policy set")
		return reconcile.Result{}, nil
	}
	// enhance logger for policy and deployment
	reqLogger = reqLogger.WithValues("policy", canaryPolicyName, "deployment", canaryInstance.Name)
	reqLogger.Info("Reconciling a Canary Deployment")

	// get the policy
	canaryPolicy := &v1.CanaryPolicy{}
	err = r.client.Get(ctx, types.NamespacedName{
		Namespace: request.Namespace,
		Name:      canaryPolicyName,
	}, canaryPolicy)
	if err != nil {
		reqLogger.Error(err, "Canary policy could not be found for deployment")
		return reconcile.Result{}, err
	}

	// reconcile the canary object for this canary deployment
	if err = r.reconcileCanaryObject(ctx, reqLogger, canaryInstance, canaryPolicy); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDeployment) reconcileCanaryObject(ctx context.Context, log logr.Logger, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) error {
	canaryConfig := &v1.Canary{}

	_, ok := canaryInstance.ObjectMeta.Labels["krane.sh/canary-config"]
	if !ok {
		log.Info("Creating Canary Config")
		canaryConfig = r.createCanaryConfig(ctx, canaryInstance, policy)
		err := r.client.Create(ctx, canaryConfig)
		if err != nil {
			// in case we already found
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "Canary Config creation failed")
				return err
			} else {
				log.Info("Canary config already found")
			}
		}

		// update the canary instance with the canary-config
		canaryInstance.ObjectMeta.Labels["krane.sh/canary-config"] = canaryConfig.Name
		log.Info("Updating canary instance reference to Canary Config")
		err = r.client.Update(ctx, canaryInstance)
		if err != nil {
			log.Error(err, "Error updating canary-config on canary instance")
			deleteErr := r.client.Delete(ctx, canaryConfig)
			if err != nil {
				log.Error(deleteErr, "Error deleting the non-updatable canary config")
			}
			return err
		}
	} else {
		log.Info("Retrieving canary config information")
		// retrieve the canary configuration for this deployment
		canaryConfigName := policy.Name + "-" + canaryInstance.Name
		log = log.WithValues("canary-config", canaryConfigName)

		err := r.client.Get(ctx, types.NamespacedName{
			Namespace: canaryInstance.Namespace,
			Name:      canaryConfigName,
		}, canaryConfig)
		if err != nil {
			log.Error(err, "Getting canary configuration failed")
			return err
		}
	}

	// check for baseline deployment
	baseline := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: canaryInstance.Namespace,
		Name:      canaryConfig.Spec.Baseline,
	}, baseline)
	if err != nil {
		if errors.IsNotFound(err) {
			baseline, err = r.createBaselineDeployment(ctx, canaryInstance, policy)
			if err != nil {
				log.Error(err, "An error occurred when creating baseline configuration")
				return err
			}
		} else {
			log.Error(err, "An error occurred when getting Baseline")
			return err
		}

		log.Info("Creating baseline deployment")
		err = r.client.Create(ctx, baseline)
		if err != nil {
			log.Error(err, "An error occurred during baseline creation")
			return err
		}
	}

	log.Info("Deployment Reconciliation complete")

	return nil
}

func (r *ReconcileDeployment) createCanaryConfig(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) *v1.Canary {
	return &v1.Canary{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policy.Name + "-" + canaryInstance.Name,
			Namespace: policy.Namespace,
		},
		Spec: v1.CanarySpec{
			Policy:   policy.Name,
			Canary:   canaryInstance.Name,
			Baseline: canaryInstance.Name + "-baseline",
			Base:     policy.Spec.Base,
		},
		Status: v1.CanaryStatus{
			Progress: v1.CanaryProgress_Initializing,
		},
	}
}

func (r *ReconcileDeployment) createBaselineDeployment(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*appsv1.Deployment, error) {
	baseline := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      canaryInstance.Name + "-baseline",
			Namespace: canaryInstance.Namespace,
			Labels: map[string]string{
				"release": "baseline",
			},
		},
	}

	// get the old deployment to retrieve containers
	base := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: canaryInstance.Namespace,
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
		baseline.Spec = *canaryInstance.Spec.DeepCopy()
		// copy over previous container configuration
		baseline.Spec.Template.Spec.Containers = base.Spec.Template.Spec.Containers

	case v1.BaselineModeOld:
		baseline.Spec = *base.Spec.DeepCopy()
	}

	// argh, golang, why you no support pointers
	singleReplica := int32(1)
	baseline.Spec.Replicas = &singleReplica

	// connect selectors
	baseline.Spec.Selector.MatchLabels["release"] = "baseline"
	baseline.Spec.Template.ObjectMeta.Labels["release"] = "baseline"

	return baseline, nil
}
