package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/iostrovok/conveyor/protobuf/go/nodes"
	"github.com/iostrovok/conveyormaster/server/messager"
)

type Server struct {
	message messager.IMessage
}

var GlobalGRPCServer *Server

// Serve starts server
func Start(addr string, message messager.IMessage) error {

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Cannot bind to %s", addr)
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(50),
	)

	GlobalGRPCServer = &Server{
		message: message,
	}

	// added health check to service
	nodes.RegisterMasterNodeServer(grpcServer, GlobalGRPCServer)

	fmt.Printf("....grpc started on %s\n", addr)
	return grpcServer.Serve(lis)
}

// UpdateNodeInfo gets node info and returns instructions
func (g *Server) UpdateNodeInfo(ctx context.Context, in *nodes.SlaveNodeInfoRequest) (*nodes.SimpleResult, error) {
	//fmt.Printf("in: %s\n", in.String())
	g.message.AddGrpcRequest(in)
	return &nodes.SimpleResult{}, nil
}
