package controller

import (
	cassandrav1alpha1 "github.com/rook/rook/pkg/apis/cassandra.rook.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Cluster is
	// synced.
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a
	// Cluster fails to sync due to a resource of the same name already
	// existing.
	ErrSyncFailed = "ErrSyncFailed"

	MessageResourceSynced = "Cluster synced successfully"

	// Messages to display when experiencing an error.
	MessageHeadlessServiceSyncFailed      = "Failed to sync Headless Service for cluster"
	MessageMemberServicesSyncFailed       = "Failed to sync MemberServices for cluster"
	MessageMemberRoleSyncFailed           = "Failed to sync MemberRole for cluster"
	MessageMemberRoleBindingSyncFailed    = "Failed to sync MemberRoleBinding for cluster"
	MessageMemberServiceAccountSyncFailed = "Failed to sync MemberServiceAccount for cluster"
	MessageUpdateStatusFailed             = "Failed to update status for cluster"
	MessageClusterSyncFailed              = "Failed to sync cluster"
)

// Sync attempts to sync the given Cassandra Cluster.
// NOTE: the Cluster Object is a DeepCopy. Modify at will.
func (cc *ClusterController) Sync(c *cassandrav1alpha1.Cluster) error {

	// Sync Headless Service for Cluster
	if err := cc.SyncClusterHeadlessService(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageHeadlessServiceSyncFailed,
		)
		return err
	}

	// Sync Cluster Member Services
	if err := cc.SyncMemberServices(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageMemberServicesSyncFailed,
		)
		return err
	}

	// Sync ServiceAccount for Cluster Members
	if err := cc.SyncMemberServiceAccount(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageMemberServiceAccountSyncFailed,
		)
		return err
	}

	// Sync Role for Cluster Members
	if err := cc.SyncMemberRole(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageMemberRoleSyncFailed,
		)
		return err
	}

	// Sync RoleBinding for Cluster Members
	if err := cc.SyncMemberRoleBinding(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageMemberRoleBindingSyncFailed,
		)
		return err
	}

	// Update Status
	if err := cc.UpdateStatus(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageUpdateStatusFailed,
		)
		return err
	}

	// Sync Cluster
	if err := cc.SyncCluster(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageClusterSyncFailed,
		)
		return err
	}

	cc.recorder.Event(
		c,
		corev1.EventTypeNormal,
		SuccessSynced,
		MessageResourceSynced,
	)

	return nil
}
