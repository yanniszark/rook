package util

import (
	"encoding/json"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/client/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

// PatchPod patches the old Pod so that it matches the
// new Pod.
func PatchPod(old, new *corev1.Pod, kubeClient kubernetes.Interface) error {

	oldJSON, err := json.Marshal(old)
	if err != nil {
		return err
	}

	newJSON, err := json.Marshal(new)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldJSON, newJSON, corev1.Pod{})
	if err != nil {
		return err
	}

	_, err = kubeClient.CoreV1().Pods(old.Namespace).Patch(old.Name, types.StrategicMergePatchType, patchBytes)
	return err
}

// PatchStatefulSet patches the old StatefulSet so that it matches the
// new StatefulSet.
func PatchStatefulSet(old, new *appsv1.StatefulSet, kubeClient kubernetes.Interface) error {

	oldJSON, err := json.Marshal(old)
	if err != nil {
		return err
	}

	newJSON, err := json.Marshal(new)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldJSON, newJSON, appsv1.StatefulSet{})
	if err != nil {
		return err
	}

	_, err = kubeClient.AppsV1().StatefulSets(old.Namespace).Patch(old.Name, types.StrategicMergePatchType, patchBytes)
	return err
}

// PatchCluster patches the old Cluster so that it matches the new Cluster.
func PatchClusterStatus(c *cassandrav1alpha1.Cluster, rookClient versioned.Interface) error {

	// JSON Patch RFC 6902
	patch := []struct {
		Op    string                          `json:"op"`
		Path  string                          `json:"path"`
		Value cassandrav1alpha1.ClusterStatus `json:"value"`
	}{
		{
			Op:    "add",
			Path:  "/status",
			Value: c.Status,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = rookClient.CassandraV1alpha1().Clusters(c.Namespace).Patch(c.Name, types.JSONPatchType, patchBytes)
	return err

}
