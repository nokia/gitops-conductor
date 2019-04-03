package reporting

import (
	"os/exec"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/nokia/gitops-conductor/pkg/apis/ops/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_reporter")

func SendReport(reporting *v1alpha1.Reporting, hash string, ops *v1alpha1.GitOps) {

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command("sh", "-c", reporting.Plugin),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Error(err, "Failed to connect plugin")
		return
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("report_grpc")
	if err != nil {
		log.Error(err, "Failed to call gRPC client")
	}

	reporter := raw.(Reporter)
	reporter.UpdateResult(hash, ops.Status.RootFolder)

}
