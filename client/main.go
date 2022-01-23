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
	addr          = flag.String("addr", "localhost:8080", "chat server address")
	reciever_id   = flag.String("reciever_id", "", "reciever_id")
	sender_id     = flag.String("sender_id", "", "sender_id")
	reciever_mode = flag.Bool("reciever_mode", true, "reciever mode")
)

func recieve(reciever_id string, chatClient *pb.ChatClient, ctx *context.Context) {
	var r pb.RecieveRequest
	r.RecieverId = []byte(reciever_id)
	recieve_client, err := (*chatClient).Recieve(*ctx, &r)
	if err != nil {
		log.Fatalf("failed to create recieve stream with error [%s]", err)
	}
	for {
		response, err := recieve_client.Recv()
		if err != nil {
			log.Fatalf("failed to recieve message with error [%s]", err)
		}
		fmt.Printf("recieved message [%s] from [%s]\n", response.GetIncommingMessage().Data, response.GetIncommingMessage().SenderId)
	}
}

func send(reciever_id string, sender_id string, chatClient *pb.ChatClient, ctx *context.Context) {
	var m pb.Message
	m.SenderId = []byte(sender_id)
	m.RecieverId = []byte(reciever_id)
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
	if (*reciever_mode && (*reciever_id == "" || *sender_id != "")) || (!*reciever_mode && (*reciever_id == "" || *sender_id == "")) {
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

	if *reciever_mode {
		recieve(*reciever_id, &chatClient, &ctx)
	} else {
		send(*reciever_id, *sender_id, &chatClient, &ctx)
	}
}
