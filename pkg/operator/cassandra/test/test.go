package test

import (
	"fmt"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/operator/cassandra/controller/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewSimpleCluster(members int32) *cassandrav1alpha1.Cluster {
	return &cassandrav1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: cassandrav1alpha1.APIVersion,
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "test-ns",
		},
		Spec: cassandrav1alpha1.ClusterSpec{
			Version: "3.1.11",
			Mode:    cassandrav1alpha1.ClusterModeCassandra,
			Datacenter: cassandrav1alpha1.DatacenterSpec{
				Name: "test-dc",
				Racks: []cassandrav1alpha1.RackSpec{
					{
						Name:    "test-rack",
						Members: members,
					},
				},
			},
		},
	}
}

// MemberServicesForCluster returns the member services for a given cluster
func MemberServicesForCluster(c *cassandrav1alpha1.Cluster) []runtime.Object {

	services := []runtime.Object{}
	for _, r := range c.Spec.Datacenter.Racks {
		for i := int32(0); i < c.Status.Racks[r.Name].Members; i++ {
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-%s-%d", c.Name, c.Spec.Datacenter.Name, r.Name, i),
					Namespace: c.Namespace,
					Labels:    util.RackLabels(r, c),
				},
			}
			services = append(services, svc)
		}
	}
	return services
}
