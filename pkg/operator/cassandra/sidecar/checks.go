package sidecar

import (
	"fmt"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	"github.com/yanniszark/go-nodetool/nodetool"
	"net/http"
)

// setupHTTPChecks brings up the liveness and readiness probes
func (m *MemberController) setupHTTPChecks() error {

	http.HandleFunc(constants.LivenessProbePath, livenessCheck(m))
	http.HandleFunc(constants.ReadinessProbePath, readinessCheck(m))

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", constants.ProbePort), nil)
	// If ListenAndServe returns, something went wrong
	m.logger.Fatalf("Error in HTTP checks: %s", err.Error())
	return err

}

func livenessCheck(m *MemberController) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		status := http.StatusOK

		// Check if JMX is reachable
		_, err := m.nodetool.Status()
		if err != nil {
			m.logger.Errorf("Liveness check failed with error: %s", err.Error())
			status = http.StatusServiceUnavailable
		}

		w.WriteHeader(status)

	}
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
			// Up means the member is alive
			if localNode.Status != nodetool.NodeStatusUp {
				return fmt.Errorf("Unexpected local node status: %s", localNode.Status)
			}
			// Check local node state
			// Normal means that the member has completed bootstrap and joined the cluster
			if localNode.State != nodetool.NodeStateNormal {
				return fmt.Errorf("Unexpected local node state: %s", localNode.State)
			}
			return nil
		}()

		if err != nil {
			m.logger.Errorf("Readiness check failed with error: %s", err.Error())
			status = http.StatusServiceUnavailable
		}

		w.WriteHeader(status)
	}

}
