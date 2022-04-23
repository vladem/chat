package main

import (
	"flag"
	"log"
	"net"
	"time"

	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/chatmanager"
	cmn "whcrc/chat/server/common"

	"net/http"
	_ "net/http/pprof"

	"github.com/google/uuid"
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
	cookie := uuid.NewString()
	go func() {
		reader := s.chatManager.GetReaderFor(chatId, cm.ReaderConfig{BufferSize: 100}, cookie)
		defer reader.Close()
		for {
			msg, _ := reader.Recv(srv.Context().Done())
			if msg == nil {
				break
			}
			srv.Send(&pb.ChatEvent{Event: &pb.ChatEvent_IncommingMessage{IncommingMessage: msg}})
		}
		log.Printf("reader [%s] subhandler done", chatId.String())
	}()
	writer := s.chatManager.GetWriterFor(chatId, cookie)
	defer writer.Close()
	for {
		req, err := srv.Recv()
		if err != nil {
			log.Printf("error on receive from grpc channel: %v", err)
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
	log.Printf("handler for %s done", chatId.String())
	return nil
}

func main() {
	go func() {
		addr := "localhost:6060"
		log.Printf("profile server listening at %s\n", addr)
		log.Println(http.ListenAndServe(addr, nil))
	}()
	log.SetFlags(log.Lmicroseconds)
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	chatManager := cm.CreateChatManager()
	go chatManager.Act()
	pb.RegisterChatServer(s, &ChatService{chatManager: chatManager})
	log.Printf("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
