package storage_test

import (
	"fmt"
	"testing"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"

	cm "whcrc/chat/server/common"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	s := storage.GetChatStorage(cm.GetChatId("test", "simple"))
	go s.Act()
	defer s.Close()
	for i := 1; i < 11; i++ {
		message := pb.Message{Data: []byte(fmt.Sprintf("blabla_%d", i))}
		resChan := s.Write(&message)
		err := <-resChan
		assert.Empty(t, err, "error on write")
	}
	for i := 1; i < 11; i++ {
		resChan := s.Read(uint64(i))
		message := <-resChan
		assert.NotEmptyf(t, message, "message with id [%d] is empty", i)
		fmt.Printf("TestSimple: received %v\n", message)
		assert.Equal(t, message.Data, []byte(fmt.Sprintf("blabla_%d", i)), "unexpected message received")
		assert.Equal(t, message.MessageId, uint64(i), "unexpected messageId received")
	}
}

func TestConcurrentWrites(t *testing.T) {
	s := storage.GetChatStorage(cm.GetChatId("test", "concurrent"))
	go s.Act()
	defer s.Close()
	writeResults := make([]chan error, 0)
	for i := 1; i < 11; i++ {
		message := pb.Message{Data: []byte(fmt.Sprintf("blabla_%d", i))}
		resChan := s.Write(&message)
		writeResults = append(writeResults, resChan)
	}
	for i, writeResChan := range writeResults {
		assert.Emptyf(t, <-writeResChan, "%d-th write operation failed", i)
	}
	for i := 1; i < 11; i++ {
		resChan := s.Read(uint64(i))
		message := <-resChan
		assert.NotEmptyf(t, message, "message with id [%d] is empty", i)
		fmt.Printf("TestConcurrentWrites: received %v\n", message)
		assert.Equal(t, message.Data, []byte(fmt.Sprintf("blabla_%d", i)), "unexpected message received")
	}
}
