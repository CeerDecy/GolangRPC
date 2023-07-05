package rpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

// GrpcServer 服务端
type GrpcServer struct {
	listen   net.Listener
	g        *grpc.Server
	register []func(g *grpc.Server)
	opt      []grpc.ServerOption
}

// NewGrpcServer 创建一个GRPC服务器
func NewGrpcServer(addr string, opts ...GrpcOption) (*GrpcServer, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	rpcServer := &GrpcServer{
		listen: listen,
	}
	for _, v := range opts {
		v.Apply(rpcServer)
	}
	server := grpc.NewServer(rpcServer.opt...)
	rpcServer.g = server
	return rpcServer, nil
}

func (s *GrpcServer) Run() error {
	for _, f := range s.register {
		f(s.g)
	}
	return s.g.Serve(s.listen)
}

func (s *GrpcServer) Stop() {
	s.g.Stop()
}

func (s *GrpcServer) Register(f func(g *grpc.Server)) {
	s.register = append(s.register, f)
}

// GrpcOption 参数
type GrpcOption interface {
	Apply(s *GrpcServer)
}

// DefaultGrpcOption 默认参数实现
type DefaultGrpcOption struct {
	f func(s *GrpcServer)
}

// Apply 应用参数
func (d *DefaultGrpcOption) Apply(s *GrpcServer) {
	d.f(s)
}

// WithGrpcOptions 返回一个GrpcOption
func WithGrpcOptions(ops ...grpc.ServerOption) GrpcOption {
	return &DefaultGrpcOption{
		f: func(s *GrpcServer) {
			s.opt = append(s.opt, ops...)
		},
	}
}

// GrpcClient GRPC客户端
type GrpcClient struct {
	Conn *grpc.ClientConn
}

func NewGrpcClient(config *GrpcClientOption) (*GrpcClient, error) {
	ctx := context.Background()
	dialOption := config.dialOptions
	if config.Block {
		if config.DialTimeout > time.Duration(0) {
			var cancelFunc context.CancelFunc
			ctx, cancelFunc = context.WithTimeout(ctx, config.DialTimeout)
			defer cancelFunc()
		}
		dialOption = append(dialOption, grpc.WithBlock())
	}
	if config.KeepAlive != nil {
		dialOption = append(dialOption, grpc.WithKeepaliveParams(*config.KeepAlive))
	}
	conn, err := grpc.DialContext(ctx, config.Address, dialOption...)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{
		Conn: conn,
	}, err
}

type GrpcClientOption struct {
	Address     string
	Block       bool
	DialTimeout time.Duration
	ReadTimeout time.Duration
	Direct      bool
	KeepAlive   *keepalive.ClientParameters
	dialOptions []grpc.DialOption
}

func DefaultGrpcClientConfig(addr string) *GrpcClientOption {
	return &GrpcClientOption{
		Address: addr,
		dialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
		DialTimeout: time.Second * 3,
		ReadTimeout: time.Second * 3,
		Block:       true,
	}
}
