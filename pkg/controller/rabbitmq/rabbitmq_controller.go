package rabbitmq

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/lesolise/rabbitmq-operator/pkg/utils"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	lesolisev1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	err = c.Watch(&source.Kind{Type: &v1beta12.Ingress{}}, &handler.EnqueueRequestForOwner{
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

	// reconcile
	for _, fun := range []reconcileFun{
		r.reconcileRabbitMQ,
		r.reconcileRabbitMQManager,
		r.reconcileRabbitMQProxy,
		r.reconcileClusterStatus,
	} {
		if err = fun(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileRabbitMQ) reconcileRabbitMQ(instance *lesolisev1.RabbitMQ) error {
	//check config map
	config := utils.NewConfigMapForCR(instance)
	// Set Kafka instance as the owner and controller
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
	} else if err != nil {
		return fmt.Errorf("GET svc fail : %s", err)
	}

	//check ingress
	rmi := utils.NewRabbitMQManagementIngressForCR(instance)
	if err := controllerutil.SetControllerReference(instance, rmi, r.scheme); err != nil {
		return fmt.Errorf("SET ingress Owner fail : %s", err)
	}
	foundKmi := &v1beta12.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rmi.Name, Namespace: rmi.Namespace}, foundKmi)

	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new rabbitmq management ingress", "Namespace", rmi.Namespace, "Name", rmi.Name)
		err = r.client.Create(context.TODO(), rmi)
		if err != nil {
			return fmt.Errorf("Create rabbitmq management ingress fail : %s", err)
		}
	} else if err != nil {
		return fmt.Errorf("GET rabbitmq management ingress fail : %s", err)
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
	} else if err != nil {
		return fmt.Errorf("GET proxy svc fail : %s", err)
	}

	return nil
}

func (r *ReconcileRabbitMQ) reconcileClusterStatus(instance *lesolisev1.RabbitMQ) error {
	return r.client.Status().Update(context.TODO(), instance)
}
