package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime"

	cockroachdbrookio "github.com/rook/rook/pkg/apis/cockroachdb.rook.io"
)

const (
	CustomResourceGroup = "cassandra.rook.io"
	Version             = "v1alpha1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: cockroachdbrookio.GroupName, Version: Version}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder and AddToScheme will stay in k8s.io/kubernetes
	SchemeBuilder runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to api.Scheme
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Cluster{},
		&ClusterList{},
		&Member{},
		&MemberList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
}