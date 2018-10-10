package util

import (
	"fmt"
	"github.com/coreos/pkg/capnslog"
	cassandrarookio "github.com/rook/rook/pkg/apis/cassandra.rook.io"
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"
	corelisters "k8s.io/client-go/listers/core/v1"
)

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "util")

// GetPodsForCluster returns the existing Pods for
// the given cluster
func GetPodsForCluster(cluster *cassandrav1alpha1.Cluster, podLister corelisters.PodLister) ([]*corev1.Pod, error) {

	clusterRequirement, err := labels.NewRequirement(constants.ClusterNameLabel, selection.Equals, []string{cluster.Name})
	if err != nil {
		logger.Errorf("Error trying to create clusterRequirement")
		return nil, err
	}
	clusterSelector := labels.NewSelector().Add(*clusterRequirement)
	return podLister.List(clusterSelector)

}

// VerifyOwner checks if the owner Object is the controller
// of the obj Object and returns an error if it isn't.
func VerifyOwner(obj, owner metav1.Object) error {
	if !metav1.IsControlledBy(obj, owner) {
		ownerRef := metav1.GetControllerOf(obj)
		return fmt.Errorf(
			"'%s/%s' is foreign owned: "+
				"it is owned by '%v', not '%s/%s'.",
			obj.GetNamespace(), obj.GetName(),
			ownerRef,
			owner.GetNamespace(), owner.GetName(),
		)
	}
	return nil
}

// NewControllerRef returns an OwnerReference to
// the provided Cluster Object
func NewControllerRef(c *cassandrav1alpha1.Cluster) metav1.OwnerReference {
	return *metav1.NewControllerRef(c, schema.GroupVersionKind{
		Group:   cassandrarookio.GroupName,
		Version: "v1alpha1",
		Kind:    "Cluster",
	})
}

// RefFromString is a helper function that takes a string
// and outputs a reference to that string.
// Useful for initializing a string pointer from a literal.
func RefFromString(s string) *string {
	return &s
}

func StatefulSetNameForRack(rackName, dcName, clusterName string) string {
	return fmt.Sprintf("%s-%s-%s", clusterName, dcName, rackName)
}

func ServiceAccountNameForMembers(c *cassandrav1alpha1.Cluster) string {
	return fmt.Sprintf("%s-member", c.Name)
}

func RoleNameForMembers(c *cassandrav1alpha1.Cluster) string {
	return fmt.Sprintf("%s-member", c.Name)
}

func RoleBindingNameForMembers(c *cassandrav1alpha1.Cluster) string {
	return fmt.Sprintf("%s-member", c.Name)
}

func HeadlessServiceNameForCluster(c *cassandrav1alpha1.Cluster) string {
	return fmt.Sprintf("%s-hs", c.Name)
}

func CassandraImageForCluster(c *cassandrav1alpha1.Cluster) string {
	repo := "cassandra"
	if c.Spec.Repository != nil {
		repo = *c.Spec.Repository
	}
	return fmt.Sprintf("%s:%s", repo, c.Spec.Version)
}

func StatefulSetForRack(r *cassandrav1alpha1.RackSpec, c *cassandrav1alpha1.Cluster, rookImage string) *appsv1.StatefulSet {

	rackLabels := RackLabels(r.Name, c)
	stsName := StatefulSetNameForRack(r.Name, c.Spec.Datacenter.Name, c.Name)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            stsName,
			Namespace:       c.Namespace,
			Labels:          rackLabels,
			OwnerReferences: []metav1.OwnerReference{NewControllerRef(c)},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &r.Members,
			// Use a common Headless Service for all StatefulSets
			ServiceName: HeadlessServiceNameForCluster(c),
			Selector: &metav1.LabelSelector{
				MatchLabels: RackLabels(r.Name, c),
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: rackLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "shared",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cassandra",
							Image:           CassandraImageForCluster(c),
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "intra-node",
									ContainerPort: 7000,
								},
								{
									Name:          "tls-intra-node",
									ContainerPort: 7001,
								},
								{
									Name:          "jmx",
									ContainerPort: 7199,
								},
								{
									Name:          "cql",
									ContainerPort: 9042,
								},
								{
									Name:          "thrift",
									ContainerPort: 9160,
								},
								{
									Name:          "jolokia",
									ContainerPort: 8778,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  constants.EnvVarSharedVolume,
									Value: constants.SharedDirName,
								},
							},
							// TODO: unprivileged entrypoint
							Command: []string{
								"/bin/bash",
								"-c",
								`set -e
                                 CONFIG="/etc/cassandra"
                                 mkdir -p "$SHARED_VOLUME"/default-temp
                                 # Copy default config files
                                 if [ ! -d "$SHARED_VOLUME"/default ]; then
                                     cp -a "$CONFIG"/* "$SHARED_VOLUME"/default-temp
                                     mv "$SHARED_VOLUME"/default-temp "$SHARED_VOLUME"/default 
                                 fi
                                 # Get custom config files
                                 while ! [ -d "$SHARED_VOLUME"/custom ]; do sleep 10; done; cp -a "$SHARED_VOLUME"/custom/* "$CONFIG"
                                 # TODO: get plugins (jolokia, jmx-exporter)
                                 # Permissions
                                 chown -R cassandra:cassandra /var/lib/cassandra /var/log/cassandra $CONFIG
                                 # Start cassandra
                                 exec gosu cassandra cassandra -f
                                 `,
							},
							Resources:    r.Resources["cassandra"],
							VolumeMounts: volumeMountsForRack(r, false),
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: int32(15),
								TimeoutSeconds:      int32(5),
								// TODO: Investigate if it's ok to call status every 10 seconds
								PeriodSeconds: int32(10),
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(constants.ReadinessProbePort),
										Path: constants.ReadinessProbePath,
									},
								},
							},
						},
						{
							Name:            "sidecar",
							Image:           rookImage,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									Name:          "readiness",
									ContainerPort: constants.ReadinessProbePort,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: constants.EnvVarPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: constants.EnvVarPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: constants.EnvVarPodIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
							},
							Args: []string{
								"cassandra",
								"sidecar",
							},
							Resources:    r.Resources["sidecar"],
							VolumeMounts: volumeMountsForRack(r, true),
						},
					},
					ServiceAccountName: ServiceAccountNameForMembers(c),
					Affinity:           affinityForRack(r),
					Tolerations:        tolerationsForRack(r),
				},
			},
			VolumeClaimTemplates: volumeClaimTemplatesForRack(r.Storage.VolumeClaimTemplates),
		},
	}
}

// TODO: Maybe move this logic to a defaulter
func volumeClaimTemplatesForRack(claims []corev1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {

	if len(claims) == 0 {
		return claims
	}

	for i := range claims {
		claims[i].Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}
	return claims
}

// TODO: Modify to handle JBOD
func volumeMountsForRack(r *cassandrav1alpha1.RackSpec, readOnly bool) []corev1.VolumeMount {

	vm := []corev1.VolumeMount{
		{
			Name:      "shared",
			MountPath: constants.SharedDirName,
		},
	}

	if len(r.Storage.VolumeClaimTemplates) > 0 {
		vm = append(vm, corev1.VolumeMount{
			Name:      r.Storage.VolumeClaimTemplates[0].Name,
			ReadOnly:  readOnly,
			MountPath: "/var/lib/cassandra/data",
		})
	}
	return vm
}

func tolerationsForRack(r *cassandrav1alpha1.RackSpec) []corev1.Toleration {

	if r.Placement == nil {
		return nil
	}
	return r.Placement.Tolerations
}

func affinityForRack(r *cassandrav1alpha1.RackSpec) *corev1.Affinity {

	if r.Placement == nil {
		return nil
	}

	return &corev1.Affinity{
		PodAffinity:     r.Placement.PodAffinity,
		PodAntiAffinity: r.Placement.PodAntiAffinity,
		NodeAffinity:    r.Placement.NodeAffinity,
	}
}
