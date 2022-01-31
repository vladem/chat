package main

import (
	"context"
	"flag"
	"log"
	"net"

	pb "whcrc/chat/proto"

	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:8080", "chat server address")
)

type ChatService struct {
	pb.UnimplementedChatServer
}

func (s *ChatService) Send(_ context.Context, _ *pb.Message) (*pb.SendResponse, error) {
	return &pb.SendResponse{}, nil
}

func (s *ChatService) Receive(request *pb.ReceiveRequest, receiveServer pb.Chat_ReceiveServer) error {
	select {
	case <-receiveServer.Context().Done():
		return nil
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &ChatService{})
	log.Printf("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
