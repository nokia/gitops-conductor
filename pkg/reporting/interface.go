package reporting

import (
	"context"

	"google.golang.org/grpc"

	plugin "github.com/hashicorp/go-plugin"
	proto "github.com/nokia/gitops-conductor/plugin/proto"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "REPORT_PLUGIN",
	MagicCookieValue: "GitOps",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"report_grpc": &ReporterGRPCPlugin{},
}

// Reporter is the interface that we're exposing as a plugin.
type Reporter interface {
	UpdateResult(githash string, gitroot string) error
}

// This is the implementation of plugin.Plugin so we can serve/consume this.
type ReporterPlugin struct {
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Reporter
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type ReporterGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Reporter
}

func (p *ReporterGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewReportClient(c)}, nil
}
