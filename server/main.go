package main

import (
	"flag"
	"log"
	"net"

	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/chatmanager"

	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:8080", "chat server address")
)

type ChatService struct {
	pb.UnimplementedChatServer
	chatManager cm.ChatManager
}

func (s *ChatService) Communicate(communicateServer pb.Chat_CommunicateServer) error {
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &ChatService{chatManager: cm.CreateChatManager()})
	log.Printf("server listening at %v\n", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
