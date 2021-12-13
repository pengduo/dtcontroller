package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/cri-api/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/vmsdk"
)

// DtNodeReconciler reconciles a DtNode object
type DtNodeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.dtwave.com,resources=dtnodes/finalizers,verbs=update

func (dtnodeReconciler *DtNodeReconciler) Reconcile(ctx context.Context,
	req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	dtnode := &appsv1.DtNode{}

	err := dtnodeReconciler.Client.Get(context.TODO(), req.NamespacedName, dtnode)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	ip := dtnode.Spec.Ip
	username := dtnode.Spec.User
	password := dtnode.Spec.Password
	vURL := strings.Join([]string{"https://", username, ":", password, "@", ip, "/sdk"}, "")

	dtnode.Status.Phase = "Ready"

	vmclient, err := vmsdk.Vmclient(ctx, vURL, username, password)
	if err != nil {
		fmt.Println("error when building vm client")
		dtnode.Status.Phase = "UnReady"
		return ctrl.Result{}, nil
	}

	//更新status字段
	version := vmclient.Version
	dtnode.Status.Version = version

	createTime := dtnode.ObjectMeta.CreationTimestamp
	currentTime := time.Now()
	age := currentTime.Local().UTC().Sub(createTime.Time)
	dtnode.Status.Age = strconv.FormatFloat(age.Hours(), 'f', 2, 64)

	vmsdk.GetDtNodeInfo(ctx, vmclient.Client, dtnode)

	vms := vmsdk.GetVms(ctx, vmclient.Client)
	dtnode.Status.Vms = len(vms)

	dtnodeReconciler.Status().Update(ctx, dtnode)
	return ctrl.Result{}, nil
}

// 注册到manager
func (r *DtNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DtNode{}).
		Complete(r)
}
