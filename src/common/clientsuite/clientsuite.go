package clientsuite

import (
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

type CommonClientSuite struct {
	RegistryAddr string
	ServiceName  string
}

func (s CommonClientSuite) Options() []client.Option {
	// 服务发现
	r, err := consul.NewConsulResolver(s.RegistryAddr)
	if err != nil {
		panic(err)
	}

	opts := []client.Option{
		client.WithResolver(r),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	}

	// 配置熔断 (默认200个样本，失败率达到50%熔断，熔断10s)
	cbs := circuitbreak.NewCBSuite(func(ri rpcinfo.RPCInfo) string {
		return circuitbreak.RPCInfo2Key(ri)
	})

	opts = append(opts,
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: s.ServiceName}),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithMuxConnection(2),
		client.WithCircuitBreaker(cbs),
	)

	return opts
}
