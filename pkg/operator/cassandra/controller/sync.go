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

	MessageRackCreated  = "Rack %s created"
	MessageRackScaledUp = "Rack %s scaled up to %d members"

	// Messages to display when experiencing an error.
	MessageHeadlessServiceSyncFailed = "Failed to sync Headless Service for cluster"
	MessageMemberServicesSyncFailed  = "Failed to sync MemberServices for cluster"
	MessageUpdateStatusFailed        = "Failed to update status for cluster"
	MessageClusterSyncFailed         = "Failed to sync cluster"
)

// Sync attempts to sync the given Cassandra Cluster.
// NOTE: the Cluster Object is a DeepCopy. Modify at will.
func (cc *ClusterController) Sync(c *cassandrav1alpha1.Cluster) error {

	// Sync Headless Service for Cluster
	if err := cc.syncClusterHeadlessService(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageHeadlessServiceSyncFailed,
		)
		return err
	}

	// Sync Cluster Member Services
	if err := cc.syncMemberServices(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageMemberServicesSyncFailed,
		)
		return err
	}

	// Update Status
	if err := cc.updateStatus(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageUpdateStatusFailed,
		)
		return err
	}

	// Sync Cluster
	if err := cc.syncCluster(c); err != nil {
		cc.recorder.Event(
			c,
			corev1.EventTypeWarning,
			ErrSyncFailed,
			MessageClusterSyncFailed,
		)
		return err
	}

	return nil
}
