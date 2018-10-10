package controller

import (
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/operator/cassandra/controller/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyncMemberRole checks if the member Role exists and creates it if it doesn't
// it creates it
func (cc *ClusterController) SyncMemberRole(cluster *cassandrav1alpha1.Cluster) error {

	memberRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.RoleNameForMembers(cluster),
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster.Name),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"services", "pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	existingRole, err := cc.roleLister.Roles(memberRole.Namespace).Get(memberRole.Name)
	// If we get an error but without the IsNotFound error raised
	// then something is wrong with the network, so requeue.
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// If the service already exists, check that it's
	// controlled by the given Cluster
	if err == nil {
		return util.VerifyOwner(existingRole, cluster)
	}

	// At this point, the Role doesn't exist, so we are free to create it
	_, err = cc.kubeClient.RbacV1().Roles(memberRole.Namespace).Create(memberRole)
	return err
}

// SyncMemberRoleBinding checks if the member Role exists and if it doesn't, creates it
func (cc *ClusterController) SyncMemberRoleBinding(cluster *cassandrav1alpha1.Cluster) error {

	memberRoleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.RoleBindingNameForMembers(cluster),
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster.Name),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  "",
				Kind:      rbacv1.ServiceAccountKind,
				Name:      util.ServiceAccountNameForMembers(cluster),
				Namespace: cluster.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			// TODO: Investigate if this is included as a constant in an rbac package
			Kind: "Role",
			Name: util.ServiceAccountNameForMembers(cluster),
		},
	}

	existingRoleBinding, err := cc.roleBindingLister.
		RoleBindings(memberRoleBinding.Namespace).
		Get(memberRoleBinding.Name)
	// If we get an error but without the IsNotFound error raised
	// then something is wrong with the network, so requeue.
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// If the RoleBinding already exists, check that it's
	// controlled by the given Cluster
	if err == nil {
		return util.VerifyOwner(existingRoleBinding, cluster)
	}

	// At this point, the Role doesn't exist, so we are free to create it
	_, err = cc.kubeClient.RbacV1().RoleBindings(memberRoleBinding.Namespace).Create(memberRoleBinding)

	return err
}

// SyncMemberServiceAccount checks if the member ServiceAccount exists and
// if it doesn't, creates it
func (cc *ClusterController) SyncMemberServiceAccount(cluster *cassandrav1alpha1.Cluster) error {

	memberServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            util.ServiceAccountNameForMembers(cluster),
			Namespace:       cluster.Namespace,
			Labels:          util.ClusterLabels(cluster.Name),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
		// TODO: Check if AutomountServiceAccountToken is a security risk as it's mounted in both containers
		// TODO: If Pod uses Secrets, specify them here (eg. for secure JMX)
	}

	logger.Infof("Syncing MemberServiceAccount `%s` for Cluster `%s`", memberServiceAccount.Name, cluster.Name)
	existingServiceAccount, err := cc.serviceAccountLister.
		ServiceAccounts(memberServiceAccount.Namespace).
		Get(memberServiceAccount.Name)

	// If we get an error but without the IsNotFound error raised
	// then something is wrong with the network, so requeue.
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// If the ServiceAccount already exists, check that it's
	// controlled by the given Cluster
	if err == nil {
		return util.VerifyOwner(existingServiceAccount, cluster)
	}

	// At this point, the ServiceAccount doesn't exist, so we are free to create it
	_, err = cc.kubeClient.CoreV1().
		ServiceAccounts(memberServiceAccount.Namespace).
		Create(memberServiceAccount)
	return err

}
