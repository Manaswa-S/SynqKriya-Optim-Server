package rpcserver

import (
	context "context"
	"fmt"
	"log"
	"net"
	"optim/internal/preoptim"
	"os"

	grpc "google.golang.org/grpc"
)

func InitPreOptimRPCServer(preoptim *preoptim.PreOptim) (*grpc.Server, error) {

	port, exists := os.LookupEnv("PreOptimRPCPORT")
	if !exists {
		return nil, fmt.Errorf("PreOptimRPCPORT not found in env vars")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterPreOptimServiceServer(s, &RpcServer{preoptim: preoptim})

	log.Printf("server listening at %v", lis.Addr())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("FATAL: failed to serve: %v", err)
			return
		}
	}()

	return s, nil
}

type RpcServer struct {
	UnimplementedPreOptimServiceServer

	preoptim *preoptim.PreOptim
}

func (r *RpcServer) AddCamera(ctx context.Context, rpc *AddCameraReq) (*AddCameraResp, error) {

	err := r.preoptim.AddCamera(ctx, rpc.CameraId)
	if err != nil {
		return &AddCameraResp{
			Message: "",
		}, err
	}

	return &AddCameraResp{
		Message: "",
	}, nil
}
