package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ***************************************************************************
// IMPORTANT FOR CODE GENERATION
// If the types in this file are updated, you will need to run
// `make codegen` to generate the new types under the client/clientset folder.
// ***************************************************************************

// +genClient
// +genClient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status"`
}

//  *******************************************************
//          Types for Controller - Cluster CRD
//  *******************************************************

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cluster `json:"items"`
}

type ClusterSpec struct {
	Version    string         `json:"version,omitempty"`
	Datacenter DatacenterSpec `json:"datacenter,omitempty"`
}

type ClusterStatus struct {
	State ClusterState `json:"state,omitempty"`
}

type ClusterState string

const (
	ClusterStateCreating ClusterState = "Creating"
	ClusterStateReady    ClusterState = "Ready"
	ClusterStateError    ClusterState = "Error"
)

// DatacenterSpec defines the desired state of a Cassandra Datacenter.
// A Cassandra Datacenter contains 1 or more Racks, defined by RackSpec.

type DatacenterSpec struct {
	Name  string     `json:"name"`
	Racks []RackSpec `json:"racks"`
}

// RackSpec defines the desired state of a Cassandra Rack.
// A Cassandra Rack contains 1 or more Cassandra instances.

type RackSpec struct {
	Name      string                  `json:"name"`
	Config    string                  `json:"config,omitempty"`
	Replicas  int                     `json:"replicas"`
	Resources v1.ResourceRequirements `json:"resources"`
	// TODO: storage options to select Persistent Volumes
}

type VolumeSpec struct {
	Size             int    `json:"size"`
	StorageClassName string `json:"storageClassName"`
}

//  *******************************************************
//          Types for Sidecar - Member CRD
//  *******************************************************

// +genClient
// +genClient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Member struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Spec            MemberSpec   `json:"spec"`
	Status          MemberStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MemberList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Member `json:"items"`
}

type MemberSpec struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type MemberStatus struct {
	State MemberState `json:"state,omitempty"`
}

type MemberState string

const (
	MemberStateBootstrapping MemberState = "Bootstrapping"
	MemberStateReady         MemberState = "Ready"
)
