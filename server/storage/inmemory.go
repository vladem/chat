package storage

import (
	"errors"
	"fmt"
	"log"
	"sync"
	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/common"
)

var storages = make(map[cm.ChatId]*inMemoryChatStorage)
var storagesLock = sync.Mutex{} // todo: get rid of this lock?

func getInMemoryChatStorage(chatId cm.ChatId) ChatStorage {
	storagesLock.Lock()
	defer storagesLock.Unlock()
	var storage *inMemoryChatStorage
	ok := false
	if storage, ok = storages[chatId]; !ok {
		storage = &inMemoryChatStorage{chatId, make([]*pb.Message, 0), make(chan chatStorageAction, 100), false, make(chan bool)}
		storages[chatId] = storage
	}
	return storage
}

type inMemoryChatStorage struct {
	chatId   cm.ChatId
	messages []*pb.Message
	requests chan chatStorageAction
	acting   bool
	stopped  chan bool
}

type chatStorageAction interface {
	isChatStorageAction()
}

type actionRead struct {
	messageId uint64
	response  chan *pb.Message
}

func (a actionRead) isChatStorageAction() {
}

type actionWrite struct {
	message  *pb.Message
	response chan error
}

func (a actionWrite) isChatStorageAction() {
}

func (s *inMemoryChatStorage) Read(messageId uint64) chan *pb.Message {
	result := make(chan *pb.Message)
	s.requests <- actionRead{messageId: messageId, response: result}
	return result
}

func (s *inMemoryChatStorage) Write(message *pb.Message) chan error {
	result := make(chan error)
	s.requests <- actionWrite{message: message, response: result}
	return result
}

func (s *inMemoryChatStorage) processAction(action chatStorageAction) (proceed bool) {
	if action == nil {
		return false
	}
	switch action := action.(type) {
	default:
		log.Fatalf("invalid action %T", action)
	case actionRead:
		messageIdx := int(action.messageId) - 1
		if messageIdx < 0 || messageIdx >= len(s.messages) {
			action.response <- nil
		} else {
			action.response <- s.messages[messageIdx]
		}
	case actionWrite:
		if action.message.MessageId != 0 {
			action.response <- errors.New("MessageId != 0 on write")
		} else {
			s.messages = append(s.messages, action.message)
			action.message.MessageId = uint64(len(s.messages))
			action.response <- nil
		}
	}
	return true
}

func (s *inMemoryChatStorage) Act() {
	if s.acting {
		log.Panic("double act is forbidden")
	}
	fmt.Printf("storage for chat with id [%s] started\n", s.chatId.String())
	s.acting = true
	for s.processAction(<-s.requests) {
	}
	s.acting = false
	s.stopped <- true
	fmt.Printf("storage for chat with id [] stopped\n")
}

func (s *inMemoryChatStorage) Close() {
	s.requests <- nil
	<-s.stopped
	storagesLock.Lock()
	defer storagesLock.Unlock()
	delete(storages, s.chatId)
}
