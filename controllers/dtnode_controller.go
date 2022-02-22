package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"

	logrus "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "dtcontroller/api/v1"
	"dtcontroller/util"
	"dtcontroller/vmsdk/esxi"
)

const (
	Ready     = "ready"
	NotReady  = "notready"
	UnKnown   = "unknown"
	Connected = "connected"
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

func (r *DtNodeReconciler) Reconcile(ctx context.Context,
	req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	dtnode := &appsv1.DtNode{}
	if err := r.Get(ctx, req.NamespacedName, dtnode); err != nil {
		logrus.Info("找不到dtnode")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var specNode = dtnode.Spec.Node
	var envNode = os.Getenv("MY_NODE_NAME")
	var podName = os.Getenv("MY_POD_NAME")

	// update dtnode status
	if specNode == "" {
		logrus.Info("please set node in spec")
		dtnode.Status.Phase = NotReady
		dtnode.Status.Node = ""
		r.Status().Update(ctx, dtnode)
		return ctrl.Result{}, nil
	}
	if specNode != envNode {
		logrus.Info("pod cannot handle this dtnode, pod is : ",
			podName, ",target node is : ", specNode)
		return ctrl.Result{}, nil
	}
	logrus.Info("ok, pod is now on due", podName)
	var dtClusterMap = dtnode.Spec.DtCluster
	var dtcluster = &appsv1.DtCluster{}
	// 检查dtnode绑定的dtcluster有效性
	for key, value := range dtClusterMap {
		fmt.Println(key, value)
		if err := r.Get(ctx, client.ObjectKey{Name: key}, dtcluster); err != nil {
			logrus.Info("找不到dtcluster ", key)
			dtClusterMap[key] = UnKnown
		} else if err = checkConnect(dtcluster.Spec.Provider, dtcluster.Spec.Content); err != nil {
			logrus.Error(err)
			dtClusterMap[key] = NotReady
		} else if dtcluster.Status.Bound || dtcluster.Status.DtNode != "" {
			logrus.Info("dtcluster has been bound,", dtcluster.Name)
		} else {
			dtClusterMap[key] = Connected
			// update dtcluster
			dtcluster.Status.Bound = true
			dtcluster.Status.DtNode = dtnode.Name
			r.Status().Update(ctx, dtcluster)
		}
	}
	// update dtnode spec
	dtnode.Spec.DtCluster = dtClusterMap
	r.Update(ctx, dtnode)
	dtnode.Status.Phase = Ready
	dtnode.Status.Node = envNode
	r.Status().Update(ctx, dtnode)
	return ctrl.Result{}, nil
}

// checkContent is used to check if the dtcluster right
func checkConnect(provider string, content map[string]string) error {

	logrus.Info("检查连通性, provider = ", provider)
	if provider == "esxi" {
		var ip = content["ip"]
		var username = content["username"]
		var password = content["password"]
		logrus.Info("username = ", username)
		logrus.Info("password = ", password)
		logrus.Info("ip = ", ip)
		if ip == "" || username == "" || password == "" {
			return &util.Err{Msg: "check is there an error in ip username or password"}
		}
		return checkConnectESXI(ip, username, password)
	}
	return &util.Err{Msg: "unsupported provider"}
}

// if provider is set to esxi, check if accessful
func checkConnectESXI(ip string, username string, password string) error {
	vURL := strings.Join([]string{"https://", username, ":", password, "@", ip, "/sdk"}, "")
	_, err := esxi.Vmclient(context.Background(), vURL, username, password)
	if err != nil {
		fmt.Println("error when building vm client")
		return err
	}
	logrus.Info("connect to esxi success", ip)
	return nil
}

// if provider is set to aliyun, check if accessful
// todo
func checkAliyun() error {

	return nil
}

// 注册到manager
func (r *DtNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DtNode{}).
		Complete(r)
}
