package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	pb "whcrc/chat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr       = flag.String("addr", "localhost:8080", "chat server address")
	receiverId = flag.String("receiver_id", "", "receiver_id")
	senderId   = flag.String("sender_id", "", "sender_id")
)

func printReceived(stream pb.Chat_CommunicateClient) {
	for {
		event, err := stream.Recv()
		if err != nil {
			log.Fatalf("on recv %v", err)
		}
		msg := event.GetIncommingMessage()
		if msg == nil {
			log.Fatalf("expected message, got: %v", event)
		}
		fmt.Printf("%v\n", msg)
	}
}

func main() {
	flag.Parse()
	if *receiverId == "" || *senderId == "" {
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
	stream, err := chatClient.Communicate(ctx)
	if err != nil {
		log.Fatalf("on stream creation %v", err)
	}

	{
		err = stream.Send(&pb.ClientEvent{
			Event: &pb.ClientEvent_CommunicateParams{
				CommunicateParams: &pb.CommunicateParams{
					SenderId:   []byte(*senderId),
					ReceiverId: []byte(*receiverId),
				},
			},
		})
		if err != nil {
			log.Fatalf("on send to server: %v", err)
		}
	}

	go printReceived(stream)

	stdinReader := bufio.NewReader(os.Stdin)
	for {
		text, err := stdinReader.ReadString('\n')
		if err != nil {
			log.Fatalf("on stdin read: %v", err)
		}
		err = stream.Send(&pb.ClientEvent{
			Event: &pb.ClientEvent_OutgoingMessage{
				OutgoingMessage: text,
			},
		})
		if err != nil {
			log.Fatalf("on send to server: %v", err)
		}
	}
}
