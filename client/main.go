package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	pb "whcrc/chat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr         = flag.String("addr", "localhost:8080", "chat server address")
	receiverId   = flag.String("receiver_id", "", "receiver_id")
	senderId     = flag.String("sender_id", "", "sender_id")
	receiverMode = flag.Bool("receiver_mode", true, "receiver mode")
)

func receive(receiverId string, chatClient *pb.ChatClient, ctx *context.Context) {
	var r pb.ReceiveRequest
	r.ReceiverId = []byte(receiverId)
	receiveClient, err := (*chatClient).Receive(*ctx, &r)
	if err != nil {
		log.Fatalf("failed to create receive stream with error [%s]", err)
	}
	for {
		response, err := receiveClient.Recv()
		if err != nil {
			log.Fatalf("failed to receive message with error [%s]", err)
		}
		fmt.Printf("received message [%s] from [%s]\n", response.GetIncommingMessage().Data, response.GetIncommingMessage().SenderId)
	}
}

func send(receiverId string, senderId string, chatClient *pb.ChatClient, ctx *context.Context) {
	var m pb.Message
	m.SenderId = []byte(senderId)
	m.ReceiverId = []byte(receiverId)
	m.Data = []byte("privet👀")
	m.MessageId = 1

	_, err := (*chatClient).Send(*ctx, &m)
	if err != nil {
		log.Fatalf("failed to send request with error [%s]", err)
	}
	fmt.Printf("send done succesfully")
}

func main() {
	flag.Parse()
	if (*receiverMode && (*receiverId == "" || *senderId != "")) || (!*receiverMode && (*receiverId == "" || *senderId == "")) {
		log.Fatalf("missing args")
	}
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect with error [%s]", err)
	}
	defer conn.Close()

	chatClient := pb.NewChatClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *receiverMode {
		receive(*receiverId, &chatClient, &ctx)
	} else {
		send(*receiverId, *senderId, &chatClient, &ctx)
	}
}
