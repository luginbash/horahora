package config

import (
	"time"

	"github.com/caarlos0/env"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	schedulerproto "github.com/horahoradev/horahora/scheduler/protocol"
	userproto "github.com/horahoradev/horahora/user_service/protocol"
	videoproto "github.com/horahoradev/horahora/video_service/protocol"

	"google.golang.org/grpc"
)

type Config struct {
	UserServiceGRPCAddress      string `env:"UserServiceGRPCAddress,required"`
	VideoServiceGRPCAddress     string `env:"VideoServiceGRPCAddress,required"`
	SchedulerServiceGRPCAddress string `env:"SchedulerServiceGRPCAddress,required"`

	VideoClient     videoproto.VideoServiceClient
	UserClient      userproto.UserServiceClient
	SchedulerClient schedulerproto.SchedulerClient
}

func New() (*Config, error) {
	config := Config{}

	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	retryCallOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(1000 * time.Millisecond)),
		grpc_retry.WithMax(7),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryCallOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryCallOpts...)),
	}

	videoGRPCConn, err := grpc.Dial(config.VideoServiceGRPCAddress, dialOpts...)
	if err != nil {
		return nil, err
	}

	userGRPCConn, err := grpc.Dial(config.UserServiceGRPCAddress, dialOpts...)
	if err != nil {
		return nil, err
	}

	schedulerGRPCConn, err := grpc.Dial(config.SchedulerServiceGRPCAddress, dialOpts...)
	if err != nil {
		return nil, err
	}

	config.SchedulerClient = schedulerproto.NewSchedulerClient(schedulerGRPCConn)
	config.UserClient = userproto.NewUserServiceClient(userGRPCConn)
	config.VideoClient = videoproto.NewVideoServiceClient(videoGRPCConn)

	return &config, nil
}
