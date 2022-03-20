package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/chatmanager"
	cmn "whcrc/chat/server/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	addr = flag.String("addr", "localhost:8080", "chat server address")
)

type ChatService struct {
	pb.UnimplementedChatServer
	chatManager cm.ChatManager
}

func (s *ChatService) Communicate(srv pb.Chat_CommunicateServer) error {
	var (
		chatId cmn.ChatId
		params *pb.CommunicateParams
	)
	{
		req, err := srv.Recv()
		if err != nil {
			return err
		}
		params = req.GetCommunicateParams()
		if params == nil {
			return status.Error(codes.InvalidArgument, "protocol violation: expected params")
		}
		chatId = cmn.GetChatId(string(params.SenderId), string(params.ReceiverId))
	}
	go func() {
		reader := s.chatManager.GetReaderFor(chatId, cm.ReaderConfig{BufferSize: 100})
		defer reader.Close()
		for {
			msg, _ := reader.Recv(srv.Context().Done())
			if msg == nil {
				break
			}
			srv.Send(&pb.ChatEvent{Event: &pb.ChatEvent_IncommingMessage{IncommingMessage: msg}})
		}
		fmt.Printf("reader [%s] subhandler done\n", chatId.String())
	}()
	writer := s.chatManager.GetWriterFor(chatId)
	defer writer.Close()
	for {
		req, err := srv.Recv()
		if err != nil {
			fmt.Printf("error on receive from grpc channel: %v\n", err)
			break
		}
		msg := req.GetOutgoingMessage()
		if msg == "" {
			return status.Error(codes.InvalidArgument, "protocol violation: expected message")
		}
		writer.Send(&pb.Message{
			Timestamp:  uint64(time.Now().UnixNano()),
			MessageId:  0,
			SenderId:   params.SenderId,
			ReceiverId: params.ReceiverId,
			Data:       []byte(msg),
		})
	}
	fmt.Printf("handler for %s done\n", chatId.String())
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	chatManager := cm.CreateChatManager()
	go chatManager.Act()
	pb.RegisterChatServer(s, &ChatService{chatManager: chatManager})
	log.Printf("server listening at %v\n", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
