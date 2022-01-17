package controllers

import (
	"context"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "dtcontroller/api/v1"
)

// MachineGroupReconciler reconciles a MachineGroup object
type MachineGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machinegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machinegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=machinegroups/finalizers,verbs=update

func (r *MachineGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var err error
	var dtnode = &appsv1.DtNode{}
	var machinegroup = &appsv1.MachineGroup{}
	if err = r.Get(ctx, req.NamespacedName, machinegroup); err != nil {
		logrus.Info("找不到{}资源", machinegroup.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dtnodeName := machinegroup.Spec.DtNode
	if err := r.Get(ctx, client.ObjectKey{Name: dtnodeName}, dtnode); err != nil {
		logrus.Info("找不到{}资源", dtnodeName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if dtnode.Status.Phase != "Ready" || machinegroup.Spec.Rs < 1 || machinegroup.Spec.Rs > 10 {
		logrus.Info("dtnode状态异常或者副本数量异常")
		machinegroup.Status.Phase = "Failed"
		machinegroup.Status.Rs = strings.Join([]string{"0", strconv.Itoa(int(machinegroup.Spec.Rs))}, "/")
		r.Status().Update(ctx, machinegroup)
		return ctrl.Result{}, nil
	}

	// todo

	machinegroup.Status.Phase = "Ready"
	machinegroup.Status.Rs = strings.Join([]string{strconv.Itoa(int(machinegroup.Spec.Rs)), strconv.Itoa(int(machinegroup.Spec.Rs))}, "/")
	r.Status().Update(ctx, machinegroup)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.MachineGroup{}).
		Complete(r)
}
