package controller

import (
	"fmt"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/operator/cassandra/controller/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	StatefulSetName = "%s-%s-%s"
)

// UpdateStatus updates the status of the given Cassandra Cluster
func (cc *ClusterController) UpdateStatus(c *cassandrav1alpha1.Cluster) error {

	clusterStatus := cassandrav1alpha1.ClusterStatus{
		Racks: map[string]cassandrav1alpha1.RackStatus{},
	}
	logger.Infof("Updating Status for cluster %s in namespace %s", c.Name, c.Namespace)

	// Update readyMembers
	for _, rack := range c.Spec.Datacenter.Racks {

		status := cassandrav1alpha1.RackStatus{}

		// Fetch corresponding StatefulSet from lister
		sts, err := cc.statefulSetLister.StatefulSets(c.Namespace).
			Get(fmt.Sprintf(StatefulSetName, c.Name, c.Spec.Datacenter.Name, rack.Name))
		// If it wasn't found, continue
		if errors.IsNotFound(err) {
			continue
		}
		// If we got a different error, requeue and log it
		if err != nil {
			return err
		}

		// Update readyMembers
		status.ReadyMembers = sts.Status.ReadyReplicas
		clusterStatus.Racks[rack.Name] = status
	}
	c.Status = clusterStatus
	return nil
}

// SyncCluster checks the Status and performs reconciliation for
// the given Cassandra Cluster.
func (cc *ClusterController) SyncCluster(c *cassandrav1alpha1.Cluster) error {

	// For each rack, check if a status entry exists
	for _, rack := range c.Spec.Datacenter.Racks {
		if _, ok := c.Status.Racks[rack.Name]; !ok {
			logger.Infof("Attempting to create Rack %s", rack.Name)
			return cc.createStatefulSet(util.StatefulSetForRack(&rack, c, cc.rookImage))
		}
		logger.Infof("Rack %s already created", rack.Name)
	}
	return nil
}

func (d *ClusterController) createStatefulSet(sts *appsv1.StatefulSet) error {
	_, err := d.kubeClient.AppsV1().StatefulSets(sts.Namespace).Create(sts)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return fmt.Errorf("Error trying to create StatefulSet %s in namespace %s : %s", sts.Name, sts.Namespace, err.Error())
}
