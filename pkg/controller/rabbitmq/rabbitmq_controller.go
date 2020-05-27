package rabbitmq

import (
	"context"
	"fmt"
	v12 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	lesolisev1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	"github.com/lesolise/rabbitmq-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_rabbitmq")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RabbitMQ Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitMQ{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RabbitMQ
	err = c.Watch(&source.Kind{Type: &lesolisev1.RabbitMQ{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner RabbitMQ
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lesolisev1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lesolisev1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lesolisev1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	//解除对ingress的监听，防止删除共用ingress？但需要考虑删除逻辑
	/*err = c.Watch(&source.Kind{Type: &v1beta12.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lesolisev1.RabbitMQ{},
	})
	if err != nil {
		return err
	}*/

	// Watch for changes to secondary resource Deployment and requeue the owner Kafka
	err = c.Watch(&source.Kind{Type: &v12.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &lesolisev1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRabbitMQ implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitMQ{}

// ReconcileRabbitMQ reconciles a RabbitMQ object
type ReconcileRabbitMQ struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	log        logr.Logger
	needUpdate bool
}

type reconcileFun func(cluster *lesolisev1.RabbitMQ) error

// Reconcile reads that state of the cluster for a RabbitMQ object and makes changes based on the state read
// and what is in the RabbitMQ.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRabbitMQ) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Namespace", request.Namespace, "Name", request.Name)
	r.log.Info("调谐中")
	r.needUpdate = false

	// Fetch the RabbitMQ instance
	instance := &lesolisev1.RabbitMQ{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//check if default values will be used
	changed := utils.CheckCR(instance)

	if changed {
		r.log.Info("设置默认配置")
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("设置默认配置失败: %s", err)
		}
		//retry reconcile
		return reconcile.Result{Requeue: true}, nil
	}

	//设置常量值
	if instance.Status.RabbitmqManagerPassword == "" {
		r.needUpdate = true
		instance.Status.RabbitmqManagerPassword = GetRandomString(16)
	}
	if instance.Status.RabbitmqManagerUsername == "" {
		r.needUpdate = true
		instance.Status.RabbitmqManagerUsername = "rmq_admin"
	}
	if instance.Status.RabbitmqUrl == "" {
		r.needUpdate = true
		instance.Status.RabbitmqUrl = "rmq-svc-" + instance.Name
	}
	if instance.Status.RabbitmqPort == "" {
		r.needUpdate = true
		instance.Status.RabbitmqPort = "5672"
	}
	if instance.Status.RabbitmqProxyUrl == "" {
		r.needUpdate = true
		instance.Status.RabbitmqProxyUrl = "rmq-mqp-svc-" + instance.Name + ":8080"
	}
	if instance.Status.RabbitmqManagerPath == "" {
		r.needUpdate = true
		instance.Status.RabbitmqManagerPath = "/" + instance.Namespace + "-" + instance.Name + "-rabbitmq/"
	}

	if instance.Status.RabbitmqManagerUrl == "" {
		r.needUpdate = true
		if instance.Spec.ManagerHostAlias == "" {
			instance.Status.RabbitmqManagerUrl = instance.Spec.ManagerHost + instance.Status.RabbitmqManagerPath
		} else {
			instance.Status.RabbitmqManagerUrl = instance.Spec.ManagerHostAlias + instance.Status.RabbitmqManagerPath
		}
	}
	//马上保存这些常量值
	r.reconcileClusterStatus(instance)

	// check ServiceAccount
	sa := utils.NewServiceAccountForCR(instance)
	foundSa := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, foundSa)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("在namespace下创建rabbitmq需要使用的ServiceAccount", "Namespace", sa.Namespace)
		err = r.client.Create(context.TODO(), sa)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		// any exception
		return reconcile.Result{}, err
	}

	// check Role
	role := utils.NewRoleForCR(instance)
	foundRole := &v1beta1.Role{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, foundRole)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("在namespace下创建rabbitmq需要使用的Role", "Namespace", role.Namespace)
		err = r.client.Create(context.TODO(), role)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		// any exception
		return reconcile.Result{}, err
	}

	// check Role Binding
	roleBinding := utils.NewRoleBindingForCR(instance)
	foundRoleBinding := &v1beta1.RoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, foundRoleBinding)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("在namespace下创建rabbitmq需要使用的RoleBinding", "Namespace", roleBinding.Namespace)
		err = r.client.Create(context.TODO(), roleBinding)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		// any exception
		return reconcile.Result{}, err
	}

	// reconcile
	for _, fun := range []reconcileFun{
		r.reconcileFinalizers,
		r.reconcileRabbitMQ,
		r.reconcileRabbitMQManager,
		r.reconcileMQManagementTools,
		r.reconcileRabbitMQProxy,
		r.reconcileServiceMonitor,
	} {
		if err = fun(instance); err != nil {
			r.reconcileClusterStatus(instance)
			return reconcile.Result{}, err
		} else {
			r.reconcileClusterStatus(instance)
		}
	}

	if instance.Status.Progress != 1.0 {
		instance.Status.Progress = 1.0
		r.needUpdate = true
		r.reconcileClusterStatus(instance)
	}
	return reconcile.Result{}, nil
}

func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func (r *ReconcileRabbitMQ) reconcileFinalizers(instance *lesolisev1.RabbitMQ) (err error) {
	r.log.Info("最终调谐")
	// instance is not deleted
	if instance.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(instance.ObjectMeta.Finalizers, utils.Finalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, utils.Finalizer)
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return err
			}
		}
		return r.cleanupOrphanPVCs(instance)
	} else {
		r.log.Info("最终调谐", "时间戳", instance.DeletionTimestamp)
		// instance is deleted
		if utils.ContainsString(instance.ObjectMeta.Finalizers, utils.Finalizer) {
			if err = r.cleanUpAllPVCs(instance); err != nil {
				return err
			}

			//删除ingress path
			foundIngress := &v1beta12.Ingress{}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: "mq-ingress", Namespace: instance.Spec.IngressNamespace}, foundIngress)

			if err != nil && errors.IsNotFound(err) {

			} else if err != nil {

			} else {
				utils.DeleteManagementPathFromIngress(instance, foundIngress)
				utils.DeleteRabbitMQToolsPathFromIngress(instance, foundIngress)
				err = r.client.Update(context.TODO(), foundIngress)
				if err != nil {
					return fmt.Errorf("删除ingress path出现问题: %s", err)
				}
			}

			instance.ObjectMeta.Finalizers = utils.RemoveString(instance.ObjectMeta.Finalizers, utils.Finalizer)
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileRabbitMQ) getPVCCount(instance *lesolisev1.RabbitMQ) (pvcCount int, err error) {
	pvcList, err := r.getPVCList(instance)
	if err != nil {
		return -1, err
	}
	pvcCount = len(pvcList.Items)
	return pvcCount, nil
}

func (r *ReconcileRabbitMQ) cleanupOrphanPVCs(instance *lesolisev1.RabbitMQ) (err error) {
	// this check should make sure we do not delete the PVCs before the STS has scaled down
	if instance.Status.Replicas == instance.Spec.Size {
		pvcCount, err := r.getPVCCount(instance)
		if err != nil {
			return err
		}
		r.log.Info("清理孤儿PVC", "当前PVC数", pvcCount, "副本数", instance.Status.Replicas)
		if pvcCount > int(instance.Spec.Size) {
			pvcList, err := r.getPVCList(instance)
			if err != nil {
				return err
			}
			for _, pvcItem := range pvcList.Items {
				// delete only Orphan PVCs
				if utils.IsPVCOrphan(pvcItem.Name, instance.Spec.Size) {
					r.deletePVC(pvcItem)
				}
			}
		}
	}
	return nil
}

func (r *ReconcileRabbitMQ) getPVCList(instance *lesolisev1.RabbitMQ) (pvList corev1.PersistentVolumeClaimList, err error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "rmq-node-" + instance.Name},
	})
	pvclistOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: selector,
	}
	pvcList := &corev1.PersistentVolumeClaimList{}
	err = r.client.List(context.TODO(), pvcList, pvclistOps)
	return *pvcList, err
}

func (r *ReconcileRabbitMQ) cleanUpAllPVCs(instance *lesolisev1.RabbitMQ) (err error) {
	pvcList, err := r.getPVCList(instance)
	if err != nil {
		return err
	}
	for _, pvcItem := range pvcList.Items {
		r.deletePVC(pvcItem)
	}
	return nil
}

func (r *ReconcileRabbitMQ) deletePVC(pvcItem corev1.PersistentVolumeClaim) {
	pvcDelete := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcItem.Name,
			Namespace: pvcItem.Namespace,
		},
	}
	r.log.Info("删除PVC", "名称", pvcItem.Name)
	err := r.client.Delete(context.TODO(), pvcDelete)
	if err != nil {
		r.log.Error(err, "删除PVC失败", "名称", pvcDelete.Name)
	}
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQ(instance *lesolisev1.RabbitMQ) error {
	r.log.Info("调谐RabbitMQ")
	//check config map
	config := utils.NewConfigMapForCR(instance)
	// Set rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, config, r.scheme); err != nil {
		return fmt.Errorf("设置ConfigMap控制器引用失败: %s", err)
	}

	// check config map
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: config.Name, Namespace: config.Namespace}, found)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建ConfigMap", "名称", config.Name)
		err = r.client.Create(context.TODO(), config)
		if err != nil {
			return fmt.Errorf("创建ConfigMap失败: %s", err)
		}
		instance.Status.Progress = 0.1
		r.needUpdate = true
	} else if err != nil {
		// any exception
		return fmt.Errorf("获取ConfigMap失败: %s", err)
	}

	//check lb svc
	lbsvc := utils.NewLBSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, lbsvc, r.scheme); err != nil {
		return fmt.Errorf("设置service控制器引用失败: %s", err)
	}
	foundLbSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: lbsvc.Name, Namespace: lbsvc.Namespace}, foundLbSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建service", "名称", lbsvc.Name)
		err = r.client.Create(context.TODO(), lbsvc)
		if err != nil {
			return fmt.Errorf("创建service失败: %s", err)
		}
		instance.Status.Progress = 0.15
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取service失败: %s", err)
	}

	//prometheus metrics
	monitorSvc := utils.NewMonitorSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, monitorSvc, r.scheme); err != nil {
		return fmt.Errorf("设置监控svc控制器引用失败: %s", err)
	}
	foundMonitorSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: monitorSvc.Name, Namespace: monitorSvc.Namespace}, foundMonitorSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建监控svc", "名称", monitorSvc.Name)
		err = r.client.Create(context.TODO(), monitorSvc)
		if err != nil {
			return fmt.Errorf("创建监控svc失败: %s", err)
		}
		instance.Status.Progress = 0.2
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取监控svc失败: %s", err)
	}

	//check sts
	sts := utils.NewStsForCR(instance)
	// Set sts as the owner and controller
	if err := controllerutil.SetControllerReference(instance, sts, r.scheme); err != nil {
		return fmt.Errorf("设置监控sts控制器引用失败: %s", err)
	}

	//check sts
	foundSts := &v1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, foundSts)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建sts", "名称", sts.Name)
		err = r.client.Create(context.TODO(), sts)
		if err != nil {
			return fmt.Errorf("创建sts失败: %s", err)
		}
	} else if err != nil {
		// any exception
		return fmt.Errorf("获取sts失败: %s", err)
	} else {
		// exists
		if foundSts.Spec.Replicas != sts.Spec.Replicas {
			utils.SyncRabbitMQSts(foundSts, sts)
			err = r.client.Update(context.TODO(), found)
			if err != nil {
				return fmt.Errorf("更新sts失败: %s", err)
			}
		}
	}

	//check rabbitmq cluster ready for use
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, foundSts)
	if err != nil {
		return fmt.Errorf("可用性检查获取sts失败: %s", err)
	}

	if foundSts.Status.ReadyReplicas != instance.Spec.Size {
		r.log.Info("RabbitMQ未就绪", "名称", sts.Name)
		instance.Status.Progress = float32(foundSts.Status.ReadyReplicas)/float32(foundSts.Status.Replicas)*0.3 + 0.2
		r.needUpdate = false
		return fmt.Errorf("RabbitMQ未就绪")
	}
	r.log.Info("RabbitMQ已就绪", "名称", sts.Name)
	instance.Status.Replicas = instance.Spec.Size

	return nil
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQManager(instance *lesolisev1.RabbitMQ) error {
	r.log.Info("调谐RabbitMQ management")
	//check svc
	svc := utils.NewManagementSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("设置management svc控制器引用失败: %s", err)
	}
	foundSvc := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建management svc", "名称", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("创建management svc失败: %s", err)
		}
		instance.Status.Progress = 0.6
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取management svc失败: %s", err)
	}

	//如果资源所在的ns 与 ingress所在的ns不同，需要额外创建ExternalName类型的svc
	if instance.Namespace != instance.Spec.IngressNamespace {
		external := utils.NewManagementExternalSvcForCR(instance)
		//关联控制
		if err := controllerutil.SetControllerReference(instance, external, r.scheme); err != nil {
			return fmt.Errorf("设置management external svc控制器引用失败: %s", err)
		}
		//检查是否已经存在
		foundExternal := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: external.Name, Namespace: external.Namespace}, foundExternal)

		if err != nil && errors.IsNotFound(err) {
			//如果不存在新建
			r.log.Info("创建management external svc", "名称", external.Name)
			err = r.client.Create(context.TODO(), external)
			if err != nil {
				return fmt.Errorf("创建management external svc失败: %s", err)
			}
		} else if err != nil {
			//如果发生错误重新调谐
			return fmt.Errorf("创建management external svc失败: %s", err)
		}
	}

	//check ingress
	rmi := utils.NewIngressForCRIfNotExists(instance)
	/*if err := controllerutil.SetControllerReference(instance, rmi, r.scheme); err != nil {
		return fmt.Errorf("SET ingress Owner fail : %s", err)
	}*/
	foundKmi := &v1beta12.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rmi.Name, Namespace: rmi.Namespace}, foundKmi)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建management ingress", "名称", rmi.Name)
		err = r.client.Create(context.TODO(), rmi)
		if err != nil {
			return fmt.Errorf("创建management ingress失败: %s", err)
		}
		instance.Status.Progress = 0.65
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取management ingress失败: %s", err)
	} else {
		utils.AppendManagementPathToIngress(instance, foundKmi)
		err = r.client.Update(context.TODO(), foundKmi)
		if err != nil {
			return fmt.Errorf("更新management ingress失败: %s", err)
		}
		instance.Status.Progress = 0.65
		r.needUpdate = false
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileMQManagementTools(instance *lesolisev1.RabbitMQ) error {
	r.log.Info("调谐管理工具")
	//check
	dep := utils.NewToolsForCR(instance)
	if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
		return fmt.Errorf("设置管理工具控制器引用失败: %s", err)
	}
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建管理工具", "名称", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return fmt.Errorf("创建管理工具失败: %s", err)
		}
		instance.Status.Progress = 0.7
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取管理工具失败: %s", err)
	}

	//check svc
	svc := utils.NewToolsSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("设置管理工具svc控制器引用失败: %s", err)
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建管理工具svc", "名称", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("创建管理工具svc失败: %s", err)
		}
		instance.Status.Progress = 0.75
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取管理工具svc失败: %s", err)
	}

	//如果资源所在的ns 与 ingress所在的ns不同，需要额外创建ExternalName类型的svc
	if instance.Namespace != instance.Spec.IngressNamespace {
		external := utils.NewToolsExternalSvcForCR(instance)
		//关联控制
		if err := controllerutil.SetControllerReference(instance, external, r.scheme); err != nil {
			return fmt.Errorf("设置管理工具external svc控制器引用失败: %s", err)
		}
		//检查是否已经存在
		foundExternal := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: external.Name, Namespace: external.Namespace}, foundExternal)

		if err != nil && errors.IsNotFound(err) {
			//如果不存在新建
			r.log.Info("创建管理工具external svc", "名称", external.Name)
			err = r.client.Create(context.TODO(), external)
			if err != nil {
				return fmt.Errorf("创建管理工具external svc失败: %s", err)
			}
		} else if err != nil {
			//如果发生错误重新调谐
			return fmt.Errorf("创建管理工具external svc失败: %s", err)
		}
	}

	//check ingress
	rmi := utils.NewIngressForCRIfNotExists(instance)
	/*if err := controllerutil.SetControllerReference(instance, rmi, r.scheme); err != nil {
		return fmt.Errorf("SET ingress Owner fail : %s", err)
	}*/
	foundKmi := &v1beta12.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rmi.Name, Namespace: rmi.Namespace}, foundKmi)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("未找到ingress", "名称", rmi.Name)
		return fmt.Errorf("未找到ingress")
	} else if err != nil {
		return fmt.Errorf("获取ingress失败: %s", err)
	} else {
		utils.AppendRabbitMQToolsPathToIngress(instance, foundKmi)
		err = r.client.Update(context.TODO(), foundKmi)
		if err != nil {
			return fmt.Errorf("追加管理工具path到ingress失败: %s", err)
		}
		instance.Status.Progress = 0.8
		r.needUpdate = false
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQProxy(instance *lesolisev1.RabbitMQ) error {
	r.log.Info("调谐代理")
	//check
	dep := utils.NewProxyForCR(instance)
	if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
		return fmt.Errorf("设置代理控制器引用失败: %s", err)
	}
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建代理", "名称", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return fmt.Errorf("创建代理失败: %s", err)
		}
		instance.Status.Progress = 0.9
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("更新代理失败: %s", err)
	}

	//check svc
	svc := utils.NewMqpSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("设置代理svc控制器引用失败: %s", err)
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建代理svc", "名称", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("创建代理svc失败: %s", err)
		}
		instance.Status.Progress = 0.95
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取代理svc失败: %s", err)
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileServiceMonitor(instance *lesolisev1.RabbitMQ) (err error) {
	r.log.Info("调谐监控")
	svcm := utils.NewSvcMonitorForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svcm, r.scheme); err != nil {
		return fmt.Errorf("设置监控service monitor控制器引用失败: %s", err)
	}
	foundSvcm := &v12.ServiceMonitor{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svcm.Name, Namespace: svcm.Namespace}, foundSvcm)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("创建监控service monitor", "名称", svcm.Name)
		err = r.client.Create(context.TODO(), svcm)
		if err != nil {
			return fmt.Errorf("创建监控service monitor失败: %s", err)
		}
		instance.Status.Progress = 1.0
		r.needUpdate = true
	} else if err != nil {
		return fmt.Errorf("获取监控service monitor失败: %s", err)
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileClusterStatus(instance *lesolisev1.RabbitMQ) error {
	if r.needUpdate {
		r.log.Info("更新状态")
		return r.client.Status().Update(context.TODO(), instance)
	} else {
		return nil
	}
}
