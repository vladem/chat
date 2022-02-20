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

type StreamHandler struct {
	service           *ChatService
	chat              cm.Chat
	chatCancel        chan bool
	readerCancel      chan bool
	clientMessages    chan *pb.ClientEvent
	incommingMessages cm.InputChannel
	outgoingMessages  cm.OutputChannel
}

func (h *StreamHandler) processClientEvent(event *pb.ClientEvent) {
	// todo:
	// 1. communicateParams:
	// a) h.chat = h.service.chatManager.GetChat(chatId)
	// b) h.incommingMessages, h.outgoingMessages = h.chat.Communicate()
	// 2. message:
	// a) verify(h.chat)
	// b) h.incommingMessages <- event
}

func (s *ChatService) Communicate(communicateServer pb.Chat_CommunicateServer) error {
	handler := StreamHandler{
		service:           s,
		chat:              nil,
		chatCancel:        make(chan bool),
		readerCancel:      make(chan bool),
		clientMessages:    make(chan *pb.ClientEvent),
		incommingMessages: nil,
		outgoingMessages:  nil,
	}
	go func() {
		// todo: read messages to handler.incommingMessages, with respect to handler.readerCancel
	}()
	for {
		select {
		case <-communicateServer.Context().Done():
			handler.chatCancel <- true
			handler.readerCancel <- true
			return nil
		case pbClientEvent := <-handler.clientMessages:
			handler.processClientEvent(pbClientEvent)
		case outgoingMessage := <-handler.outgoingMessages:
			// todo: send message to communicateServer
		}
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to [net.Listen] with error [%s]", err)
	}
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &ChatService{chatManager: cm.CreateChatManager()})
	log.Printf("server listening at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to server with error [%s]", err)
	}
}
