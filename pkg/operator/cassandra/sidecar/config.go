package sidecar

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/rook/rook/pkg/operator/cassandra/constants"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// generateConfigFiles injects the default configuration files
// with our custom values.
func (m *MemberController) generateConfigFiles() (err error) {

	logger.Info("Generating config files")

	// If the custom directory already exists, that means we already did the work
	// but we crashed. In that there is no need to generate config files.
	if _, err := os.Stat(constants.CustomConfigDirName); !os.IsNotExist(err) {
		return nil
	}

	// Poll for default config dir
	for _, err = os.Stat(constants.DefConfigDirName); err != nil; _, err = os.Stat(constants.DefConfigDirName) {
		if !os.IsNotExist(err) {
			logger.Errorf("Unexpected error while polling directory: %s", err.Error())
		}
		// Default config files haven't been copied yet.
		// Wait a little, then retry.
		time.Sleep(100 * time.Millisecond)
	}

	/////////////////////////////
	// Generate cassandra.yaml //
	/////////////////////////////

	// Read default cassandra.yaml
	config, err := ioutil.ReadFile(path.Join(constants.DefConfigDirName, "cassandra.yaml"))
	if err != nil {
		return fmt.Errorf("unexpected error trying to open cassandra.yaml: %s", err.Error())
	}

	customConfig, err := overrideConfigValues(config, m)
	if err != nil {
		return fmt.Errorf("Error trying to override config values: %s", err.Error())
	}

	// Write result to temp folder
	if err = os.Mkdir(constants.CustomConfigDirTempName, os.ModePerm); err != nil {
		logger.Errorf("error trying to create custom config dir: %s", err.Error())
		return err
	}

	if err = ioutil.WriteFile(path.Join(constants.CustomConfigDirTempName, "cassandra.yaml"), customConfig, os.ModePerm); err != nil {
		logger.Errorf("error trying to write custom cassandra.yaml: %s", err.Error())
		return err
	}

	//////////////////////////////////////////
	// Generate cassandra-rackdc.properties //
	//////////////////////////////////////////

	rackdcProperties := []byte(fmt.Sprintf("dc=%s"+"\n"+"rack=%s", m.datacenter, m.rack))
	if err = ioutil.WriteFile(path.Join(constants.CustomConfigDirTempName, "cassandra-rackdc.properties"), rackdcProperties, os.ModePerm); err != nil {
		logger.Errorf("error trying to write custom cassandra.yaml: %s", err.Error())
		return err
	}

	/////////////////////////////////////////
	//       Generate cassandra-env.sh     //
	/////////////////////////////////////////

	cassandraEnv, err := ioutil.ReadFile(path.Join(constants.DefConfigDirName, "cassandra-env.sh"))
	if err != nil {
		logger.Errorf("Error trying to open cassandra-env.sh, %s", err.Error())
		return err
	}

	jolokiaConfig := []byte(fmt.Sprintf(`JVM_OPTS="$JVM_OPTS -javaagent:%s=port=%d,host=127.0.0.1"`,
		path.Join(constants.PluginDirName, constants.JolokiaJarName), constants.JolokiaPort))

	err = ioutil.WriteFile(path.Join(constants.CustomConfigDirTempName, "/cassandra-env.sh"), append(cassandraEnv, jolokiaConfig...), os.ModePerm)
	if err != nil {
		logger.Errorf("Error trying to write cassandra-env.sh: %s", err.Error())
		return err
	}

	// Rename temp folder to the correct name
	// We do it this way because rename is an atomic operation
	if err = os.Rename(constants.CustomConfigDirTempName, constants.CustomConfigDirName); err != nil {
		logger.Errorf("error trying to write custom cassandra.yaml: %s", err.Error())
		return err
	}

	return nil
}

// overrideConfigValues overrides the default config values with
// our custom values, for the fields that are of interest to us
func overrideConfigValues(configText []byte, m *MemberController) ([]byte, error) {

	var config map[string]interface{}

	if err := yaml.Unmarshal(configText, &config); err != nil {
		return nil, fmt.Errorf("Error unmarshaling config file: %s", err.Error())
	}

	seeds, err := m.getSeeds()
	if err != nil {
		return nil, fmt.Errorf("Error getting seeds: %s", err.Error())
	}

	config["cluster_name"] = m.cluster
	delete(config, "listen_address")
	config["broadcast_address"] = m.ip
	config["rpc_address"] = "0.0.0.0"
	config["broadcast_rpc_address"] = m.ip
	config["endpoint_snitch"] = "GossipingPropertyFileSnitch"

	seedProvider := []map[string]interface{}{
		{
			"class_name": "org.apache.cassandra.locator.SimpleSeedProvider",
			"parameters": []map[string]interface{}{
				{
					"seeds": seeds,
				},
			},
		},
	}

	config["seed_provider"] = seedProvider

	return yaml.Marshal(config)
}

func copyCassandraPlugins() error {

	// First make the plugins folder
	err := os.Mkdir(constants.PluginDirName, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		logger.Errorf("Error trying to create plugins folder: %s", err.Error())
		return err
	}

	// Copy the Jolokia jar in the plugins folder
	copyCmd := exec.Command("cp", path.Join("/", constants.JolokiaJarName), path.Join(constants.PluginDirName, constants.JolokiaJarName))
	err = copyCmd.Run()
	if err != nil {
		logger.Errorf("Error trying to copy jolokia in plugins folder: %s", err.Error())
		return err
	}

	// TODO: cassandra-exporter plugin

	return nil
}

func (m *MemberController) getSeeds() (string, error) {

	var services *corev1.ServiceList
	var err error

	logger.Infof("Attempting to find seeds.")
	sel := fmt.Sprintf("%s,%s=%s", constants.SeedLabel, constants.ClusterNameLabel, m.cluster)

	for {

		services, err = m.kubeClient.CoreV1().Services(m.namespace).List(metav1.ListOptions{LabelSelector: sel})
		if err != nil {
			return "", err
		}
		if len(services.Items) > 0 {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	seeds := []string{}
	for _, svc := range services.Items {
		seeds = append(seeds, svc.Spec.ClusterIP)
	}
	return strings.Join(seeds, ","), nil
}

// Merge YAMLs merges two arbitrary YAML structures on the top level.
func mergeYAMLs(initialYAML, overrideYAML []byte) ([]byte, error) {

	var initial, override map[string]interface{}
	yaml.Unmarshal(initialYAML, &initial)
	yaml.Unmarshal(overrideYAML, &override)

	if initial == nil {
		initial = make(map[string]interface{})
	}
	// Overwrite the values onto initial
	for k, v := range override {
		initial[k] = v
	}
	return yaml.Marshal(initial)

}
