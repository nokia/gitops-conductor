package reporting

import (
	"github.com/nokia/gitops-conductor/plugin"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client plugin.ReporterPlugin  }

func (m *GRPCClient) UpdateResult(githash string) error {
	_, err := m.client.UpdateResult(context.Background(), &proto.PutRequest{
				Key:   key,
						Value: value,
							
	})
		return err

}



// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
		// This is the real implementation
			Impl Reporter
}

func (m *GRPCServer) Put(
		ctx context.Context,
			req *proto.PutRequest
		) (*proto.Empty, error) {
				return &proto.Empty{}, m.Impl.Put(req.Key, req.Value)

		}

		func (m *GRPCServer) Get(
				ctx context.Context,
					req *proto.GetRequest
				) (*proto.GetResponse, error) {
						v, err := m.Impl.Get(req.Key)
							return &proto.GetResponse{Value: v}, err

				}

