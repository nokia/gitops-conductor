package reporting

import (
	"context"

	"github.com/nokia/gitops-conductor/plugin/proto"
)

// GRPCClient is an implementation of Reporting that talks over RPC.
type GRPCClient struct{ client proto.ReportClient }

func (m *GRPCClient) UpdateResult(githash string, gitroot string) error {
	_, err := m.client.GitUpdate(context.Background(), &proto.UpdateResult{
		Githash:    "",
		GitRootDir: gitroot,
	})
	return err

}
