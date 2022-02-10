package controllers

import (
	"context"

	logrus "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "dtcontroller/api/v1"
)

// DtClusterReconciler reconciles a DtCluster object
type DtClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DtCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DtClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var dtCluster = &appsv1.DtCluster{}
	if err := r.Get(ctx, req.NamespacedName, dtCluster); err != nil {
		logrus.Info("找不到dtCluster")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dtCluster.Status.Bound = true
	r.Status().Update(ctx, dtCluster)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DtClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DtCluster{}).
		Complete(r)
}
