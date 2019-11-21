package rabbitmq

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	rabbitmqv1alpha1 "rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"
	"rabbitmq-operator/pkg/clients/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// Add creates a new Rabbitmq Controller and adds it to the Manager. The Manager will set fields on the Control    ler
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	reconcileRabbitmq := &ReconcileRabbitmq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
	reconcileRabbitmq.Create()
	return reconcileRabbitmq
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rabbitmq
	err = c.Watch(&source.Kind{Type: &rabbitmqv1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Rabbitmq
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1alpha1.Rabbitmq{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRabbitmq implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitmq{}

// ReconcileRabbitmq reconciles a Rabbitmq object
type ReconcileRabbitmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	queue  chan interface{}
}

// Reconcile reads that state of the cluster for a Rabbitmq object and makes changes based on the state read
// and what is in the Rabbitmq.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq")

	// Fetch the Rabbitmq instance
	instance := &rabbitmqv1alpha1.Rabbitmq{}
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

	config, err := rest.InClusterConfig()
	if err != nil {
		reqLogger.Error(err, "find a config error:", err.Error())
		return reconcile.Result{}, err
	}

	clientset, err := versioned.NewForConfig(config)

	if err != nil {
		reqLogger.Error(err, "find a clientset error:", err.Error())
		return reconcile.Result{}, err
	}

	var namespace string
	namespace = "default"

	// Define new Service object
	rabbitmqService := newRabbitmqService(instance)

	if err := controllerutil.SetControllerReference(instance, rabbitmqService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	//foundService := &corev1.Service{}
	foundRS := &corev1.Service{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rabbitmqService.Name, Namespace: namespace}, foundRS)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new RabbitmqService", "Service.Namespace", rabbitmqService.Namespace, "Service.Name", rabbitmqService.Name)
		err = r.client.Create(context.TODO(), rabbitmqService)
		if err != nil {
			reqLogger.Info("Creating rabbitmqService fail", "Service.Namespace", rabbitmqService.Namespace, "Service.Name", rabbitmqService.Name)
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		reqLogger.Info("Creating rabbitmqService success", "Service.Namespace", rabbitmqService.Namespace, "Service.Name", rabbitmqService.Name)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	//Define a new configmap object
	cm := newConfigMap(instance)

	// Set Rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this configmap already exists
	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new configmap", "cm.Namespace", cm.Namespace, "cm.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		reqLogger.Info("Creating cm success", "cm.Namespace", cm.Namespace, "cm.Name", cm.Name)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	//Define a new PVC object
	pvc := newPVC(instance, reqLogger)

	// Set Rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pvc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PVC already exists
	foundPVC := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}, foundPVC)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new pvc", "pvc.Namespace", pvc.Namespace, "pvc.Name", pvc.Name)
		err = r.client.Create(context.TODO(), pvc)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		reqLogger.Info("Creating pvc success", "pvc.Namespace", pvc.Namespace, "pvc.Name", pvc.Name)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Statefulset object
	statefulset := newStatefulSet(instance, reqLogger)

	// Set Rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, statefulset, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Statefulset already exists
	foundSFS := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}, foundSFS)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new statefulset", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
		err = r.client.Create(context.TODO(), statefulset)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		reqLogger.Info("Creating statefulset success", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		var statefulsetOld *appsv1.StatefulSet = &appsv1.StatefulSet{}
		statefulsetOld, err = clientset.AppV1().StatefulSets(instance.Spec.NameSpace).Get(instance.Spec.Name+"rabbitmq", metav1.GetOptions{})
		if statefulsetOld != nil && statefulsetOld.Spec.Replicas != statefulset.Spec.Replicas {
			reqLogger.Info("Updating statefulset", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)

			err = r.client.Update(context.TODO(), statefulset)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: statefulset already exists", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
	return reconcile.Result{}, nil
}

//使用工作队列相关
func (r *ReconcileRabbitmq) Create() {
	go func() {
		for {
			req := <-r.queue

			_ = req
		}
	}()
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
