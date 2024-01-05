package auth

import (
	"context"
	grpcv1 "github.com/Yankeegohome/protos/gen/go/gRPC-S"
	"google.golang.org/grpc"
)

type serverAPI struct {
	grpcv1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	grpcv1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *grpcv1.LoginRequest,
) (*grpcv1.LoginResponse, error) {
	return &grpcv1.LoginResponse{
		Token: req.GetEmail(),
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *grpcv1.RegisterRequest,
) (*grpcv1.RegisterResponse, error) {
	panic("implement me")
}
func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *grpcv1.IsAdminRequest,
) (*grpcv1.IsAdminResponse, error) {
	panic("implement me")
}
