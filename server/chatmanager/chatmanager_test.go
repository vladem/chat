package chatmanager

import (
	"fmt"
	"testing"
	pb "whcrc/chat/proto"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	chatId := ChatId{
		SenderId:   "whcrc",
		ReceiverId: "dude",
	}
	manager := CreateChatManager()
	go manager.Act()
	writer := manager.GetWriterFor(chatId)
	fmt.Printf("got writer\n")
	defer writer.Close()
	reader := manager.GetReaderFor(chatId, ReaderConfig{bufferSize: 1})
	fmt.Printf("got reader\n")
	defer reader.Close()

	sent := pb.Message{
		Timestamp:  1645960438,
		MessageId:  0,
		SenderId:   []byte("whcrc"),
		ReceiverId: []byte("dude"),
		Data:       []byte("some sophisticated piece of text"),
	}
	writer.Send(&sent)
	fmt.Printf("sent message\n")
	cancel := make(chan bool)
	received, err := reader.Recv(cancel)
	assert.Empty(t, err)
	fmt.Printf("received message: [%v]\n", received)
}
