package rabbitmq

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"
	rabbitmqv1alpha1 "rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac1 "k8s.io/api/rbac/v1"
	storage1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

// Add creates a new Rabbitmq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitmq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	var namespace string
	namespace = "default"

	// Define new Service object
	service := newService(instance)
	rabbitmqService := newRabbitmqService(instance)

	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(instance, rabbitmqService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundService := &corev1.Service{}
	foundRS := &corev1.Service{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", namespace)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			reqLogger.Info("Creating Service fail", "Service.Namespace", service.Namespace, "Service.Name", namespace+":aaa:"+err.Error())
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		reqLogger.Info("Creating Service success", "Service.Namespace", service.Namespace, "Service.Name", namespace)
	} else if err != nil {
		return reconcile.Result{}, err
	}

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

	//// Define a new PV object
	//pv := newPV(instance)
	//
	//// Set Rabbitmq instance as the owner and controller
	//if err := controllerutil.SetControllerReference(instance, pv, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}
	//
	//// Check if this PV already exists
	//foundPV := &corev1.PersistentVolume{}
	//err = r.client.Get(context.TODO(), types.NamespacedName{Name: pv.Name, Namespace: pv.Namespace}, foundPV)
	//if err != nil && errors.IsNotFound(err) {
	//	reqLogger.Info("Creating a new pv", "pv.Namespace", pv.Namespace, "pv.Name", pv.Name)
	//	err = r.client.Create(context.TODO(), pv)
	//	if err != nil {
	//		return reconcile.Result{}, err
	//	}
	//
	//	// Pod created successfully - don't requeue
	//	reqLogger.Info("Creating pv success", "pv.Namespace", pv.Namespace, "pv.Name", pv.Name)
	//} else if err != nil {
	//	return reconcile.Result{}, err
	//}

	// Define a new PV object
	statefulset := newStatefulSet(instance)

	// Set Rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, statefulset, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PV already exists
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
	}

	// Define a new PV object
	pvc := newPVC(instance)

	// Set Rabbitmq instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pvc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PV already exists
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

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: statefulset already exists", "statefulset.Namespace", statefulset.Namespace, "statefulset.Name", statefulset.Name)
	return reconcile.Result{}, nil
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

func newService(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-op",
			Labels:    map[string]string{"app": "rabbitmq"},
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 5672,
					Name: "amqp",
				},
			},
			Selector: map[string]string{"app": "rabbitmq"},
		},
	}
}

func newRabbitmqService(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-service-op",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:     15672,
					Name:     "test1",
					Protocol: corev1.ProtocolTCP,
					NodePort: 32001,
				},
				corev1.ServicePort{
					Port:     5672,
					Name:     "test2",
					Protocol: corev1.ProtocolTCP,
					NodePort: 32002,
				},
			},
			Selector: map[string]string{"app": "rabbitmq"},
		},
	}
}

func newStatefulSet(cr *rabbitmqv1alpha1.Rabbitmq) *appsv1.StatefulSet {

	//set metadata -> label
	alabels := map[string]string{"app": "rabbitmq"}
	//set ImagePullSecret
	var name corev1.LocalObjectReference
	name.Name = "regsecret"
	secrets := []corev1.LocalObjectReference{name}
	//set container
	var container corev1.Container
	container.Name = "rabbitmq"
	container.Image = cr.Spec.Image
	container.ImagePullPolicy = corev1.PullIfNotPresent
	limits := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU: resource.Quantity{
			Format: "256Mi",
		},
		corev1.ResourceMemory: resource.Quantity{
			Format: "150M",
		},
	}
	requests := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU: resource.Quantity{
			Format: "512Mi",
		},
		corev1.ResourceMemory: resource.Quantity{
			Format: "150M",
		},
	}
	container.Resources.Limits = limits
	container.Resources.Requests = requests
	container.VolumeMounts = []corev1.VolumeMount{corev1.VolumeMount{
		Name:      "rabbitmq-data",
		MountPath: "/var/lib/rabbitmq/mnesia",
	}}
	container.Env = cr.Spec.Envs
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-op",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "rabbitmq",
			Replicas:    &cr.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: alabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "rabbitmq",
					ImagePullSecrets:   secrets,
					Containers:         []corev1.Container{container},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "rabbitmq-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "rabbitmq-data-claim"},
							},
						},
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: alabels,
			},
		},
	}
}

func newServiceAccount(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-op",
			Namespace: "default",
		},
	}
}

func newPV(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rabbitmq-data-op",
			Labels: map[string]string{
				"release": "rabbitmq-data",
			},
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceStorage: resource.Quantity{
					Format: "2Gi",
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: "managed-nfs-storage",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: "/home/k8s/nfs/data/pv001",
					Path:   "10.90.101.73",
				},
			},
		},
	}
}

func newPVC(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.PersistentVolumeClaim {
	var scn string
	scn = "managed-nfs-storage"
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-data-claim-op",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: &scn,
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.Quantity{
						Format: "2Gi",
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"release": "rabbitmq-data"},
			},
		},
	}
}

func newStorageClass(cr *rabbitmqv1alpha1.Rabbitmq) *storage1.StorageClass {

	return &storage1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "managed-nfs-storage",
			Namespace: "default",
		},
		Provisioner: "fuseim.pri/ifs",
		Parameters:  map[string]string{"archiveOnDelete": "false"},
	}
}

func newRole(cr *rabbitmqv1alpha1.Rabbitmq) *rbac1.Role {
	return &rbac1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint-reader",
			Namespace: "default",
		},
		Rules: []rbac1.PolicyRule{
			rbac1.PolicyRule{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
			},
		},
	}
}

func newRoleBing(cr *rabbitmqv1alpha1.Rabbitmq) *rbac1.RoleBinding {
	return &rbac1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "endpoint-reader",
		},
		Subjects: []rbac1.Subject{
			rbac1.Subject{
				Kind: "ServiceAccount",
				Name: "rabbitmq",
			},
		},
		RoleRef: rbac1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "endpoint-reader",
		},
	}
}
