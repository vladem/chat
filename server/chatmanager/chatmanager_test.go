package chatmanager

import (
	"fmt"
	"testing"
	pb "whcrc/chat/proto"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var (
	simpleMessage = pb.Message{
		Timestamp:  1645960438,
		MessageId:  0,
		SenderId:   []byte("whcrc"),
		ReceiverId: []byte("dude"),
		Data:       []byte("some sophisticated piece of text"),
	}
	simpleReaderConfig = ReaderConfig{bufferSize: 1}
)

func TestSimple(t *testing.T) {
	chatId := ChatId{
		SenderId:   "whcrc",
		ReceiverId: "simple",
	}
	manager := CreateChatManager()
	go manager.Act()
	defer manager.Close()
	writer := manager.GetWriterFor(chatId)
	fmt.Printf("got writer\n")
	defer writer.Close()
	reader := manager.GetReaderFor(chatId, simpleReaderConfig)
	fmt.Printf("got reader\n")
	defer reader.Close()

	sent := proto.Clone(&simpleMessage).(*pb.Message)
	writer.Send(sent)
	fmt.Printf("sent message\n")
	cancel := make(chan bool)
	received, err := reader.Recv(cancel)
	assert.Empty(t, err)
	assert.Equal(t, sent.MessageId, uint64(1))
	fmt.Printf("received message: [%v]\n", received)
}

func TestManyReaders(t *testing.T) {
	chatId := ChatId{
		SenderId:   "whcrc",
		ReceiverId: "many",
	}
	manager := CreateChatManager()
	go manager.Act()
	defer manager.Close()
	var readers []ChatReader
	for i := 0; i < 10; i++ {
		readers = append(readers, manager.GetReaderFor(chatId, simpleReaderConfig))
		defer readers[len(readers)-1].Close()
	}
	writer := manager.GetWriterFor(chatId)
	defer writer.Close()

	sent := proto.Clone(&simpleMessage).(*pb.Message)
	writer.Send(sent)
	assert.Equal(t, sent.MessageId, uint64(1))
	cancel := make(chan bool)
	for i, reader := range readers {
		received, err := reader.Recv(cancel)
		assert.Empty(t, err)
		fmt.Printf("reader [%d] received message: [%v]\n", i, received)
	}
}

func TestReaderRecover(t *testing.T) {
	// cause reader's buffer to overflow; see overflow counter incremented; see all messages received eventually
	chatId := ChatId{
		SenderId:   "whcrc",
		ReceiverId: "recover",
	}
	manager := CreateChatManager()
	go manager.Act()
	defer manager.Close()
	writer := manager.GetWriterFor(chatId)
	defer writer.Close()

	readerBufferSize := uint64(3)
	messagesCount := uint64(readerBufferSize + 5)
	reader := manager.GetReaderFor(chatId, ReaderConfig{bufferSize: readerBufferSize})
	defer reader.Close()

	var sent []*pb.Message
	for i := uint64(0); i < messagesCount; i++ {
		sent = append(sent, proto.Clone(&simpleMessage).(*pb.Message))
		sent[len(sent)-1].Data = append(sent[len(sent)-1].Data, []byte(fmt.Sprintf("_%d", i))...)
		writer.Send(sent[len(sent)-1])
		assert.Equal(t, sent[len(sent)-1].MessageId, uint64(i+1))
	}

	cancel := make(chan bool)
	for i := uint64(0); i < messagesCount; i++ {
		received, err := reader.Recv(cancel)
		fmt.Printf("received message: [%v]\n", received)
		assert.Empty(t, err)
		assert.Equal(t, received.MessageId, uint64(i+1))
		if i >= readerBufferSize {
			assert.Equal(t, uint64(1), reader.GetCounters().bufferOverflow)
		}
	}
}
