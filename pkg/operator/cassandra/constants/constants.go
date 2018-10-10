package constants

// Labels
const (
	ClusterNameLabel    = "cassandra.rook.io/cluster"
	DatacenterNameLabel = "cassandra.rook.io/datacenter"
	RackNameLabel       = "cassandra.rook.io/rack"
	SeedLabel           = "cassandra.rook.io/seed"
)

// Environment Variable Names
const (
	EnvVarPodIP        = "POD_IP"
	EnvVarPodName      = "POD_NAME"
	EnvVarPodNamespace = "POD_NAMESPACE"
	EnvVarSharedVolume = "SHARED_VOLUME"
)

// Configuration Values
const (
	SharedDirName           = "/mnt/shared"
	DefConfigDirName        = SharedDirName + "/" + "default"
	CustomConfigDirName     = SharedDirName + "/" + "custom"
	CustomConfigDirTempName = CustomConfigDirName + "-" + "temp"
	PluginDirName           = SharedDirName + "/" + "plugins"

	JolokiaJarName = "jolokia.jar"
	JolokiaPort    = 8778
	JolokiaContext = "jolokia"

	ReadinessProbePort = 8080
	ReadinessProbePath = "/healthz"
)
