package plugins

import (
	"log"
	"net"

	pb "birdactyl-panel-backend/internal/plugins/proto"

	"google.golang.org/grpc"
)

var grpcServer *grpc.Server

func StartServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	grpcServer = grpc.NewServer()
	pb.RegisterPanelServiceServer(grpcServer, NewPanelServer())

	log.Printf("[plugins] gRPC server starting on %s", address)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("[plugins] gRPC server error: %v", err)
		}
	}()
	return nil
}

func StopServer() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}
}
