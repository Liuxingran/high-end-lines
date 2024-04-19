package main

import (
	"fmt"
	"google.golang.org/grpc"
	pb "high-end-lines/internal/grpc/proto/proto_go"
	"high-end-lines/internal/service"
	"high-end-lines/internal/service/middlewire"
	"net"
)

func main() {
	//args := os.Args
	//internal.DubboAst2Grpc(args[1], args[2], args[3], args[4], args[5])
	//args := os.Args
	//internal.TransDobboService2Grpc(args[1])

	//http_post.Start()
	//
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(middlewire.OrderUnaryServerInterceptor),
		grpc.ChainStreamInterceptor(middlewire.OrderStreamServerInterceptor, middlewire.OrderDoubleStreamServerInterceptor),
	)

	pb.RegisterSearchServiceServer(s, &service.SearchServiceImpl{})
	pb.RegisterOrderManagementServer(s, &service.OrderServiceImpl{})

	lis, err := net.Listen("tcp", ":8009")
	if err != nil {
		fmt.Println("net.Listen err", err)
		panic(err)
	}

	if err := s.Serve(lis); err != nil {
		fmt.Println("Serve err", err)
		panic(err)
	}
}
