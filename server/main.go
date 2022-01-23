package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	pb "whcrc/chat/proto"

	"google.golang.org/grpc"
)

var (
	addr = flag.String("addr", "localhost:8080", "chat server address")
)

type ChatService struct {
	pb.UnimplementedChatServer
	events_by_reciever map[string]chan *pb.RecieveResponse // multiple consumers are unsupported now
	events_lock        sync.Mutex
}

func (s *ChatService) getChannel(reciever_id string) chan *pb.RecieveResponse {
	s.events_lock.Lock()
	locked := true
	defer func() {
		if locked {
			s.events_lock.Unlock()
		}
	}()
	events_chan, ok := s.events_by_reciever[reciever_id]
	if !ok {
		events_chan = make(chan *pb.RecieveResponse, 100)
		s.events_by_reciever[string(reciever_id)] = events_chan
	}
	s.events_lock.Unlock()
	locked = false
	return events_chan
}

func (s *ChatService) Send(_ context.Context, message *pb.Message) (*pb.SendResponse, error) {
	events_chan := s.getChannel(string(message.RecieverId))
	events_chan <- &pb.RecieveResponse{
		Event: &pb.RecieveResponse_IncommingMessage{
			IncommingMessage: message,
		},
	}
	return &pb.SendResponse{}, nil
}

func (s *ChatService) Recieve(request *pb.RecieveRequest, recieve_server pb.Chat_RecieveServer) error {
	fmt.Printf("stream for [%s] started\n", request.RecieverId)
	defer fmt.Printf("stream for [%s] done\n", request.RecieverId)
	events_chan := s.getChannel(string(request.RecieverId))
	for event := range events_chan {
		err := recieve_server.Send(event)
		if err != nil {
			fmt.Printf("failed to send message to stream\n")
			break
		}
	}

	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &ChatService{events_by_reciever: make(map[string]chan *pb.RecieveResponse)})
	log.Printf("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
