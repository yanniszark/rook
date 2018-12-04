package sidecar

import (
	"fmt"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	"github.com/rook/rook/pkg/operator/cassandra/controller/util"
	"github.com/yanniszark/go-nodetool/nodetool"
	"k8s.io/api/core/v1"
)

func (m *MemberController) Sync(memberService *v1.Service) error {

	// Check if member must decommission
	if decommission, ok := memberService.Labels[constants.DecommissionLabel]; ok {
		// Check if member has already decommissioned
		if decommission == constants.LabelValueTrue {
			return nil
		}
		// Else, decommission member
		if err := m.nodetool.Decommission(); err != nil {
			m.logger.Errorf("Error during decommission: %s", err.Error())
		}
		// Confirm memberService has been decommissioned
		if opMode, err := m.nodetool.OperationMode(); err != nil || opMode != nodetool.NodeOperationModeDecommissioned {
			return fmt.Errorf("error during decommission, operation mode: %s, error: %v", opMode, err)
		}
		// Update Label
		old := memberService.DeepCopy()
		memberService.Labels[constants.DecommissionLabel] = constants.LabelValueTrue
		if err := util.PatchService(old, memberService, m.kubeClient); err != nil {
			return fmt.Errorf("error patching MemberService, %s", err.Error())
		}

	}

	return nil
}
