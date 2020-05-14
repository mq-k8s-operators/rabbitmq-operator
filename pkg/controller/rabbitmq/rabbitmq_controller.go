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

	return nil
}

// blank assignment to verify that ReconcileRabbitMQ implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitMQ{}

// ReconcileRabbitMQ reconciles a RabbitMQ object
type ReconcileRabbitMQ struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
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
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.log.Info("Reconciling RabbitMQ")

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
		r.log.Info("Setting default settings for RabbitMQ")
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, fmt.Errorf("Setting default fail : %s", err)
		}
		//retry reconcile
		return reconcile.Result{Requeue: true}, nil
	}

	if instance.Status.RabbitmqManagerPassword == "" {
		instance.Status.RabbitmqManagerPassword = GetRandomString(16)
	}
	if instance.Status.RabbitmqManagerUsername == "" {
		instance.Status.RabbitmqManagerUsername = "rmq_admin"
	}
	if instance.Status.RabbitmqUrl == "" {
		instance.Status.RabbitmqUrl = "rmq-svc-" + instance.Name
	}
	if instance.Status.RabbitmqPort == "" {
		instance.Status.RabbitmqPort = "5672"
	}
	if instance.Status.RabbitmqProxyUrl == "" {
		instance.Status.RabbitmqProxyUrl = "rmq-mqp-svc-" + instance.Name + ":8080"
	}
	if instance.Status.RabbitmqManagerPath == "" {
		instance.Status.RabbitmqManagerPath = "/" + instance.Namespace + "-" + instance.Name + "-rabbitmq/"
	}

	if instance.Status.RabbitmqManagerUrl == "" {
		if instance.Spec.ManagerHostAlias == "" {
			instance.Status.RabbitmqManagerUrl = instance.Spec.ManagerHost + instance.Status.RabbitmqManagerPath
		} else {
			instance.Status.RabbitmqManagerUrl = instance.Spec.ManagerHostAlias + instance.Status.RabbitmqManagerPath
		}
	}
	r.reconcileClusterStatus(instance)

	// check ServiceAccount
	sa := utils.NewServiceAccountForCR(instance)
	foundSa := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, foundSa)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating ServiceAccount for Namespace", sa.Namespace)
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
		r.log.Info("Creating Role for Namespace", role.Namespace)
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
		r.log.Info("Creating RoleBinding for Namespace", roleBinding.Namespace)
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
			r.log.Info("reconcileClusterStatus with error")
			r.reconcileClusterStatus(instance)
			return reconcile.Result{}, err
		} else {
			r.log.Info("reconcileClusterStatus without error")
			r.reconcileClusterStatus(instance)
		}
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
	r.log.Info("instance.DeletionTimestamp is ", instance.DeletionTimestamp)
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
					return fmt.Errorf("update ingress fail when reconcileFinalizers: %s", err)
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
		r.log.Info("cleanupOrphanPVCs", "PVC Count", pvcCount, "ReadyReplicas Count", instance.Status.Replicas)
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
	r.log.Info("Deleting PVC", "With Name", pvcItem.Name)
	err := r.client.Delete(context.TODO(), pvcDelete)
	if err != nil {
		r.log.Error(err, "Error deleteing PVC.", "Name", pvcDelete.Name)
	}
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQ(instance *lesolisev1.RabbitMQ) error {
	//check config map
	config := utils.NewConfigMapForCR(instance)
	// Set rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, config, r.scheme); err != nil {
		return fmt.Errorf("SET ConfigMap Owner fail : %s", err)
	}

	// check config map
	found := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: config.Name, Namespace: config.Namespace}, found)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", config.Namespace, "ConfigMap.Name", config.Name)
		err = r.client.Create(context.TODO(), config)
		if err != nil {
			return fmt.Errorf("Create ConfigMap fail : %s", err)
		}
		instance.Status.Progress = 0.1
	} else if err != nil {
		// any exception
		return fmt.Errorf("GET ConfigMap fail : %s", err)
	}

	//check lb svc
	lbsvc := utils.NewLBSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, lbsvc, r.scheme); err != nil {
		return fmt.Errorf("SET SVC Owner fail : %s", err)
	}
	foundLbSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: lbsvc.Name, Namespace: lbsvc.Namespace}, foundLbSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new lb svc", "Svc.Namespace", lbsvc.Namespace, "Svc.Name", lbsvc.Name)
		err = r.client.Create(context.TODO(), lbsvc)
		if err != nil {
			return fmt.Errorf("Create lb svc fail : %s", err)
		}
		instance.Status.Progress = 0.2
	} else if err != nil {
		return fmt.Errorf("GET svc fail : %s", err)
	}

	//prometheus metrics
	monitorSvc := utils.NewMonitorSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, monitorSvc, r.scheme); err != nil {
		return fmt.Errorf("SET SVC Owner fail : %s", err)
	}
	foundMonitorSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: monitorSvc.Name, Namespace: monitorSvc.Namespace}, foundMonitorSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new lb svc", "Svc.Namespace", monitorSvc.Namespace, "Svc.Name", monitorSvc.Name)
		err = r.client.Create(context.TODO(), monitorSvc)
		if err != nil {
			return fmt.Errorf("Create lb svc fail : %s", err)
		}
		instance.Status.Progress = 0.2
	} else if err != nil {
		return fmt.Errorf("GET svc fail : %s", err)
	}

	//check sts
	sts := utils.NewStsForCR(instance)
	// Set sts as the owner and controller
	if err := controllerutil.SetControllerReference(instance, sts, r.scheme); err != nil {
		return fmt.Errorf("SET RabbitMQ STS Owner fail : %s", err)
	}

	//check sts
	foundSts := &v1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, foundSts)
	// if not exists
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Sts", "Sts.Namespace", sts.Namespace, "Sts.Name", sts.Name)
		err = r.client.Create(context.TODO(), sts)
		if err != nil {
			return fmt.Errorf("Create sts fail : %s", err)
		}
	} else if err != nil {
		// any exception
		return fmt.Errorf("GET sts fail : %s", err)
	} else {
		// exists
		utils.SyncRabbitMQSts(foundSts, sts)
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			return fmt.Errorf("Update ZK Fail : %s", err)
		}
	}

	//check rabbitmq cluster ready for use
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, foundSts)
	if err != nil {
		return fmt.Errorf("CHECK rabbitmq Status Fail : %s", err)
	}

	if foundSts.Status.ReadyReplicas != instance.Spec.Size {
		r.log.Info("rabbitmq Not Ready", "Namespace", sts.Namespace, "Name", sts.Name)
		instance.Status.Progress = float32(foundSts.Status.ReadyReplicas)/float32(foundSts.Status.Replicas)*0.3 + 0.2
		return fmt.Errorf("rabbitmq Not Ready")
	}
	r.log.Info("rabbitmq Ready", "Namespace", sts.Namespace, "Name", sts.Name, "found", found)
	instance.Status.Replicas = instance.Spec.Size

	return nil
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQManager(instance *lesolisev1.RabbitMQ) error {
	//check svc
	svc := utils.NewManagementSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("SET Management SVC Owner fail : %s", err)
	}
	foundSvc := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Management svc", "Svc.Namespace", svc.Namespace, "Svc.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("Create headless svc fail : %s", err)
		}
		instance.Status.Progress = 0.6
	} else if err != nil {
		return fmt.Errorf("GET svc fail : %s", err)
	}

	//如果资源所在的ns 与 ingress所在的ns不同，需要额外创建ExternalName类型的svc
	if instance.Namespace != instance.Spec.IngressNamespace {
		external := utils.NewManagementExternalSvcForCR(instance)
		//关联控制
		if err := controllerutil.SetControllerReference(instance, external, r.scheme); err != nil {
			return fmt.Errorf("SET Management external svc Owner fail : %s", err)
		}
		//检查是否已经存在
		foundExternal := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: external.Name, Namespace: external.Namespace}, foundExternal)

		if err != nil && errors.IsNotFound(err) {
			//如果不存在新建
			r.log.Info("Creating a new Management external svc", "Namespace", external.Namespace, "Name", external.Name)
			err = r.client.Create(context.TODO(), external)
			if err != nil {
				return fmt.Errorf("Create Management external svc fail : %s", err)
			}
		} else if err != nil {
			//如果发生错误重新调谐
			return fmt.Errorf("GET Management external svc fail : %s", err)
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
		r.log.Info("Creating a new rabbitmq management ingress", "Namespace", rmi.Namespace, "Name", rmi.Name)
		err = r.client.Create(context.TODO(), rmi)
		if err != nil {
			return fmt.Errorf("Create rabbitmq management ingress fail : %s", err)
		}
		instance.Status.Progress = 0.65
	} else if err != nil {
		return fmt.Errorf("GET rabbitmq management ingress fail : %s", err)
	} else {
		utils.AppendManagementPathToIngress(instance, foundKmi)
		err = r.client.Update(context.TODO(), foundKmi)
		if err != nil {
			return fmt.Errorf("update rabbitmq management ingress fail : %s", err)
		}
		instance.Status.Progress = 0.65
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileMQManagementTools(instance *lesolisev1.RabbitMQ) error {
	//check
	dep := utils.NewToolsForCR(instance)
	if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
		return fmt.Errorf("SET proxy Owner fail : %s", err)
	}
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new MQManagementTools", "Namespace", dep.Namespace, "Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return fmt.Errorf("Create proxy fail : %s", err)
		}
		instance.Status.Progress = 0.7
	} else if err != nil {
		return fmt.Errorf("GET proxy fail : %s", err)
	}

	//check svc
	svc := utils.NewToolsSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("SET Management SVC Owner fail : %s", err)
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new MQManagementTools svc", "Svc.Namespace", svc.Namespace, "Svc.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("Create headless svc fail : %s", err)
		}
		instance.Status.Progress = 0.75
	} else if err != nil {
		return fmt.Errorf("GET svc fail : %s", err)
	}

	//如果资源所在的ns 与 ingress所在的ns不同，需要额外创建ExternalName类型的svc
	if instance.Namespace != instance.Spec.IngressNamespace {
		external := utils.NewToolsExternalSvcForCR(instance)
		//关联控制
		if err := controllerutil.SetControllerReference(instance, external, r.scheme); err != nil {
			return fmt.Errorf("SET MQManagementTools tools external svc Owner fail : %s", err)
		}
		//检查是否已经存在
		foundExternal := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: external.Name, Namespace: external.Namespace}, foundExternal)

		if err != nil && errors.IsNotFound(err) {
			//如果不存在新建
			r.log.Info("Creating a new MQManagementTools external svc", "Namespace", external.Namespace, "Name", external.Name)
			err = r.client.Create(context.TODO(), external)
			if err != nil {
				return fmt.Errorf("Create MQManagementTools external svc fail : %s", err)
			}
		} else if err != nil {
			//如果发生错误重新调谐
			return fmt.Errorf("GET MQManagementTools external svc fail : %s", err)
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
		r.log.Info("Missing ingress", "Namespace", rmi.Namespace, "Name", rmi.Name)
		return fmt.Errorf("Missing ingress")
	} else if err != nil {
		return fmt.Errorf("GET rabbitmq management ingress fail : %s", err)
	} else {
		utils.AppendRabbitMQToolsPathToIngress(instance, foundKmi)
		err = r.client.Update(context.TODO(), foundKmi)
		if err != nil {
			return fmt.Errorf("update rabbitmq manager ingress fail : %s", err)
		}
		instance.Status.Progress = 0.8
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQProxy(instance *lesolisev1.RabbitMQ) error {
	//check
	dep := utils.NewProxyForCR(instance)
	if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
		return fmt.Errorf("SET proxy Owner fail : %s", err)
	}
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Proxy", "Namespace", dep.Namespace, "Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return fmt.Errorf("Create proxy fail : %s", err)
		}
		instance.Status.Progress = 0.9
	} else if err != nil {
		return fmt.Errorf("GET proxy fail : %s", err)
	}

	//check svc
	svc := utils.NewMqpSvcForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return fmt.Errorf("SET mqp svc Owner fail : %s", err)
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating proxy svc", "Namespace", svc.Namespace, "Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return fmt.Errorf("Create proxy svc fail : %s", err)
		}
		instance.Status.Progress = 1.0
	} else if err != nil {
		return fmt.Errorf("GET proxy svc fail : %s", err)
	}

	instance.Status.Progress = 1.0
	return nil
}

func (r *ReconcileRabbitMQ) reconcileServiceMonitor(instance *lesolisev1.RabbitMQ) (err error) {
	svcm := utils.NewSvcMonitorForCR(instance)
	if err := controllerutil.SetControllerReference(instance, svcm, r.scheme); err != nil {
		return fmt.Errorf("SET svcm Owner fail : %s", err)
	}
	foundSvcm := &v12.ServiceMonitor{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: foundSvcm.Name, Namespace: foundSvcm.Namespace}, foundSvcm)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating exporter svc", "Namespace", svcm.Namespace, "Name", svcm.Name)
		err = r.client.Create(context.TODO(), svcm)
		if err != nil {
			return fmt.Errorf("Create svcm fail : %s", err)
		}
		instance.Status.Progress = 1.0
	} else if err != nil {
		return fmt.Errorf("GET svcm fail : %s", err)
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileClusterStatus(instance *lesolisev1.RabbitMQ) error {
	return r.client.Status().Update(context.TODO(), instance)
}
