package sidecar

import (
	"fmt"
	"github.com/coreos/pkg/capnslog"
	rookClientset "github.com/rook/rook/pkg/client/clientset/versioned"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	"github.com/rook/rook/pkg/operator/cassandra/nodetool"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"net/url"
	"time"
)

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "sidecar")

// MemberController encapsulates all the tools the sidecar needs to
// talk to the Kubernetes API
type MemberController struct {
	name, namespace, ip       string
	cluster, datacenter, rack string
	kubeClient                kubernetes.Interface
	rookClient                rookClientset.Interface

	nodetool nodetool.Interface

	queue workqueue.RateLimitingInterface
}

// New return a new MemberController
func New(
	name, namespace string,
	kubeClient kubernetes.Interface,
	rookClient rookClientset.Interface,
) (*MemberController, error) {

	// Retrieve the service that we use as a static IP
	idService, err := kubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	for err != nil || idService.Spec.ClusterIP == "" {
		logger.Infof("Something went wrong trying to get Member Service %s", err.Error())
		time.Sleep(500 * time.Millisecond)
		idService, err = kubeClient.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	}

	ip := idService.Spec.ClusterIP
	logger.Infof("Got ip %s", ip)

	// Get the Member's metadata from the Pod's labels
	pod, err := kubeClient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Create a new nodetool interface to talk to Cassandra
	url, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d/jolokia/", constants.JolokiaPort))
	if err != nil {
		return nil, err
	}
	nodetool := nodetool.NewFromURL(url)

	return &MemberController{
		name:       name,
		namespace:  namespace,
		ip:         ip,
		cluster:    pod.Labels[constants.ClusterNameLabel],
		datacenter: pod.Labels[constants.DatacenterNameLabel],
		rack:       pod.Labels[constants.RackNameLabel],
		kubeClient: kubeClient,
		rookClient: rookClient,
		nodetool:   nodetool,
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}, nil
}

// Run starts executing the control loop for the sidecar
func (m *MemberController) Run(threadiness int, stopCh <-chan struct{}) error {

	defer runtime.HandleCrash()

	if err := m.onStartup(); err != nil {
		return err
	}

	logger.Infof("Main event loop")

	<-stopCh
	logger.Info("Shutting down sidecar.")
	return nil

}

func (m *MemberController) onStartup() error {

	// Setup HTTP checks
	logger.Info("Setting up HTTP Checks...")
	go func() {
		err := m.setupHTTPChecks()
		logger.Fatalf("Error with HTTP Server: %s", err.Error())
		panic("Something went wrong with the HTTP Checks")
	}()

	// Copy plugins to shared directory
	logger.Info("Copying plugins...")
	err := copyCassandraPlugins()
	if err != nil {
		return err
	}

	// Prepare config files for Cassandra
	logger.Infof("Generating cassandra config files...")
	err = m.generateConfigFiles()

	return err
}

// setupHTTPChecks brings up the liveness and readiness probes
func (m *MemberController) setupHTTPChecks() error {

	http.HandleFunc(constants.ReadinessProbePath, readinessCheck(m))
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", constants.ReadinessProbePort), nil)
	// If ListenAndServe returns, something went wrong
	logger.Fatalf("Error in HTTP checks: %s", err.Error())
	return err

}

func readinessCheck(m *MemberController) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		status := http.StatusOK

		err := func() error {
			// Contact Cassandra to learn about the status of the member
			HostIDMap, err := m.nodetool.Status()
			if err != nil {
				return fmt.Errorf("Error while executing nodetool status in readiness check: %s", err.Error())
			}
			// Get local node through static ip
			localNode, ok := HostIDMap[m.ip]
			if !ok {
				return fmt.Errorf("Couldn't find node with ip %s in nodetool status.", m.ip)
			}
			// Check local node status
			if localNode.Status != nodetool.NodeStatusUp {
				return fmt.Errorf("Unexpected local node status: %s", localNode.Status)
			}
			// Check local node state
			if localNode.State != nodetool.NodeStateNormal {
				return fmt.Errorf("Unexpected local node state: %s", localNode.State)
			}
			return nil
		}()

		if err != nil {
			logger.Errorf("Readiness check failed with error: %s", err.Error())
			status = http.StatusServiceUnavailable
		}

		w.WriteHeader(status)
	}

}
