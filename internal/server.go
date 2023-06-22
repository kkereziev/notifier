package internal

import (
	"context"
	"fmt"
	"net"

	"github.com/kkereziev/notifier/internal/config"
	pb "github.com/kkereziev/notifier/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Instance is a gRPC server.
type Instance struct {
	logger     *zap.SugaredLogger
	listener   net.Listener
	grpcServer *grpc.Server
}

// NewInstance is a constructor function for Instance, which creates new gRPC server.
func NewInstance(
	config *config.Config,
	logger *zap.SugaredLogger,
	handler pb.NotificationServiceServer,
	options []grpc.ServerOption,
) (*Instance, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port))
	if err != nil {
		return nil, err
	}

	server := &Instance{
		logger:   logger,
		listener: listener,
	}

	server.grpcServer = grpc.NewServer(serverOptions(config)...)

	server.grpcServer.RegisterService(&pb.NotificationService_ServiceDesc, handler)

	reflection.Register(server.grpcServer)

	return server, nil
}

// Start will start the gRPC server.
func (s *Instance) Start(ctx context.Context) error {
	s.logger.Info(ctx, "[gRPC] Web server is starting on ", s.listener.Addr())

	return s.grpcServer.Serve(s.listener)
}

// Stop will gracefully stop the gRPC server.
func (s *Instance) Stop(ctx context.Context) {
	s.logger.Info(ctx, "[gRPC] Web server is shutting down")

	done := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		s.logger.Warn(ctx, "[gRPC] Web server graceful-stop timeout exceeded")
		s.grpcServer.Stop()
	case <-done:
		s.logger.Info(ctx, "[gRPC] Web server gracefully stopped")
	}
}

func serverOptions(config *config.Config) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.Server.KeepAlivePolicy.Time,
			PermitWithoutStream: config.Server.KeepAlivePolicy.PermitWithoutStream,
		}),
	}
}
