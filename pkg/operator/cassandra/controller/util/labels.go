package util

import (
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// ClusterLabels returns a map of label keys and values
// for the given Cluster.
func ClusterLabels(name string) map[string]string {
	return map[string]string{
		constants.ClusterNameLabel: name,
	}
}

func DatacenterLabels(dcName string, c *cassandrav1alpha1.Cluster) map[string]string {
	dcLabels := ClusterLabels(c.Name)
	dcLabels[constants.DatacenterNameLabel] = dcName
	return dcLabels
}

func RackLabels(rackName string, c *cassandrav1alpha1.Cluster) map[string]string {
	rackLabels := DatacenterLabels(c.Spec.Datacenter.Name, c)
	rackLabels[constants.RackNameLabel] = rackName
	return rackLabels
}

// StatefulSetPodLabel returns a map of labels to uniquely
// identify a StatefulSet Pod with the given name
func StatefulSetPodLabel(name string) map[string]string {
	return map[string]string{
		appsv1.StatefulSetPodNameLabel: name,
	}
}

func RackSelector(rackName string, c *cassandrav1alpha1.Cluster) (*labels.Selector, error) {

	sel := labels.NewSelector()
	rackLabels := RackLabels(rackName, c)

	for key, val := range rackLabels {
		req, err := labels.NewRequirement(key, selection.Equals, []string{val})
		if err != nil {
			return nil, err
		}
		sel.Add(*req)
	}
	return &sel, nil
}
