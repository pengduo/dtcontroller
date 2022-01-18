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
	"dtcontroller/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// 生成一组machine

	for i := 0; i < int(machinegroup.Spec.Rs); i++ {
		var suffix string = util.String(8)
		var name string = machinegroup.Name + suffix
		machine := &appsv1.Machine{
			TypeMeta: metav1.TypeMeta{APIVersion: appsv1.GroupVersion.Version, Kind: "Machine"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: machinegroup.Namespace,
			},
			Spec: appsv1.MachineSpec{
				DtNode:   machinegroup.Spec.DtNode,
				Type:     machinegroup.Spec.Type,
				User:     machinegroup.Spec.User,
				Password: machinegroup.Spec.Password,
				Cpu:      machinegroup.Spec.Cpu,
				Memory:   machinegroup.Spec.Memory,
				Disk:     machinegroup.Spec.Disk,
				Base:     machinegroup.Spec.Base,
				Os:       machinegroup.Spec.Os,
			},
		}
		if err := assignMachine(ctx, machine, *dtnode); err != nil {
			logrus.Info("出错了", err)
		}
	}

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
