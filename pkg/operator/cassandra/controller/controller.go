package controller

import (
	"fmt"
	"github.com/coreos/pkg/capnslog"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	rookClientset "github.com/rook/rook/pkg/client/clientset/versioned"
	rookScheme "github.com/rook/rook/pkg/client/clientset/versioned/scheme"
	informersv1alpha1 "github.com/rook/rook/pkg/client/informers/externalversions/cassandra.rook.io/v1alpha1"
	listersv1alpha1 "github.com/rook/rook/pkg/client/listers/cassandra.rook.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	rbacinformers "k8s.io/client-go/informers/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	rbaclisters "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const (
	controllerName   = "cassandra-controller"
	clusterQueueName = "cluster-queue"
)

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "cassandra-controller")

// ClusterController encapsulates all the tools the controller needs
// in order to talk to the Kubernetes API
type ClusterController struct {
	rookImage                  string
	kubeClient                 kubernetes.Interface
	rookClient                 rookClientset.Interface
	clusterLister              listersv1alpha1.ClusterLister
	clusterListerSynced        cache.InformerSynced
	statefulSetLister          appslisters.StatefulSetLister
	statefulSetListerSynced    cache.InformerSynced
	serviceLister              corelisters.ServiceLister
	serviceListerSynced        cache.InformerSynced
	podLister                  corelisters.PodLister
	podListerSynced            cache.InformerSynced
	roleLister                 rbaclisters.RoleLister
	roleListerSynced           cache.InformerSynced
	roleBindingLister          rbaclisters.RoleBindingLister
	roleBindingListerSynced    cache.InformerSynced
	serviceAccountLister       corelisters.ServiceAccountLister
	serviceAccountListerSynced cache.InformerSynced

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	queue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the Kubernetes API
	recorder record.EventRecorder
}

// New returns a new ClusterController
func New(
	rookImage string,
	kubeClient kubernetes.Interface,
	rookClient rookClientset.Interface,
	clusterInformer informersv1alpha1.ClusterInformer,
	statefulSetInformer appsinformers.StatefulSetInformer,
	serviceInformer coreinformers.ServiceInformer,
	podInformer coreinformers.PodInformer,
	serviceAccountInformer coreinformers.ServiceAccountInformer,
	roleInformer rbacinformers.RoleInformer,
	roleBindingInformer rbacinformers.RoleBindingInformer,

) *ClusterController {

	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	rookScheme.AddToScheme(scheme.Scheme)
	// Create event broadcaster
	logger.Infof("creating event broadcaster...")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})

	cc := &ClusterController{
		rookImage:  rookImage,
		kubeClient: kubeClient,
		rookClient: rookClient,

		clusterLister:              clusterInformer.Lister(),
		clusterListerSynced:        clusterInformer.Informer().HasSynced,
		statefulSetLister:          statefulSetInformer.Lister(),
		statefulSetListerSynced:    statefulSetInformer.Informer().HasSynced,
		podLister:                  podInformer.Lister(),
		podListerSynced:            podInformer.Informer().HasSynced,
		serviceLister:              serviceInformer.Lister(),
		serviceListerSynced:        serviceInformer.Informer().HasSynced,
		serviceAccountLister:       serviceAccountInformer.Lister(),
		serviceAccountListerSynced: serviceAccountInformer.Informer().HasSynced,
		roleLister:                 roleInformer.Lister(),
		roleListerSynced:           roleInformer.Informer().HasSynced,
		roleBindingLister:          roleBindingInformer.Lister(),
		roleBindingListerSynced:    roleBindingInformer.Informer().HasSynced,

		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), clusterQueueName),
		recorder: recorder,
	}

	// Add event handling functions

	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newCluster := obj.(*cassandrav1alpha1.Cluster)
			cc.enqueueCluster(newCluster)
		},
		UpdateFunc: func(old, new interface{}) {
			newCluster := new.(*cassandrav1alpha1.Cluster)
			oldCluster := old.(*cassandrav1alpha1.Cluster)
			if newCluster.ResourceVersion == oldCluster.ResourceVersion {
				return
			}
			cc.enqueueCluster(newCluster)
		},
		DeleteFunc: func(obj interface{}) {
			// TODO: handle deletion
		},
	})

	statefulSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cc.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newStatefulSet := new.(*appsv1.StatefulSet)
			oldStatefulSet := old.(*appsv1.StatefulSet)
			if newStatefulSet.ResourceVersion == oldStatefulSet.ResourceVersion {
				return
			}
			cc.handleObject(new)
		},
		DeleteFunc: cc.handleObject,
	})

	return cc
}

// Run starts the ClusterController process loop
func (cc *ClusterController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer cc.queue.ShutDown()

	// 	Start the informer factories to begin populating the informer caches
	logger.Infof("starting cassandra controller")

	// Wait for the caches to be synced before starting workers
	logger.Infof("waiting for informers caches to sync...")
	if ok := cache.WaitForCacheSync(
		stopCh,
		cc.clusterListerSynced,
		cc.statefulSetListerSynced,
		cc.podListerSynced,
		cc.serviceListerSynced,
		cc.serviceAccountListerSynced,
		cc.roleListerSynced,
		cc.roleBindingListerSynced,
	); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Infof("starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(cc.runWorker, time.Second, stopCh)
	}

	logger.Infof("started workers")
	<-stopCh
	logger.Infof("Shutting down cassandra controller workers")

	return nil
}

func (cc *ClusterController) runWorker() {
	for cc.processNextWorkItem() {
	}
}

func (cc *ClusterController) processNextWorkItem() bool {
	obj, shutdown := cc.queue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer cc.queue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			cc.queue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in queue but got %#v", obj))
		}
		if err := cc.syncHandler(key); err != nil {
			cc.queue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s', requeueing: %s", key, err.Error())
		}
		cc.queue.Forget(obj)
		logger.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Cluster
// resource with the current status of the resource.
func (cc *ClusterController) syncHandler(key string) error {

	// Convert the namespace/name string into a distinct namespace and name.
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Cluster resource with this namespace/name
	cluster, err := cc.clusterLister.Clusters(namespace).Get(name)
	if err != nil {
		// The Cluster resource may no longer exist, in which case we stop processing.
		if apierrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("cluster '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	logger.Infof("handling cluster object: %+v", cluster)
	// Deepcopy here to ensure nobody messes with the cache.
	return cc.Sync(cluster.DeepCopy())
}

// enqueueCluster takes a Cluster resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Cluster.
func (cc *ClusterController) enqueueCluster(obj *cassandrav1alpha1.Cluster) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	cc.queue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Cluster resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Cluster resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (cc *ClusterController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		logger.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	logger.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// TODO: If this object is not owned by a Cluster, we should not do anything more with it.
		if ownerRef.Kind != "Cluster" {
			return
		}

		cluster, err := cc.clusterLister.Clusters(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			logger.Infof("ignoring orphaned object '%s' of cluster '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		cc.enqueueCluster(cluster)
		return
	}
}
