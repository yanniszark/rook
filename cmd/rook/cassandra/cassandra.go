package cassandra

import (
	"github.com/coreos/pkg/capnslog"
	"github.com/rook/rook/cmd/rook/rook"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/k8sutil"
	"github.com/rook/rook/pkg/util/exec"
	"github.com/spf13/cobra"
)

// Cmd exports cobra command according to the cobra documentation.
var Cmd = &cobra.Command{
	Use:    "cassandra",
	Short:  "Main command for cassandra controller pod.",
	Hidden: true,
}

var (
	logger = capnslog.NewPackageLogger("github.com/rook/rook", "cassandracmd")
)

func init() {
	Cmd.AddCommand(operatorCmd)
	Cmd.AddCommand(sidecarCmd)
}

func createContext() *clusterd.Context {
	executor := &exec.CommandExecutor{}
	return &clusterd.Context{
		Executor:  executor,
		ConfigDir: k8sutil.DataDir,
		LogLevel:  rook.Cfg.LogLevel,
	}
}
