package storage

import (
	"errors"
	"log"
	"sync"
	pb "whcrc/chat/proto"
)

var storages = make(map[string]*inMemoryChatStorage)
var storagesLock = sync.Mutex{} // todo: get rid of this lock?

func getInMemoryChatStorage(chatId string) ChatStorage {
	storagesLock.Lock()
	defer storagesLock.Unlock()
	var storage *inMemoryChatStorage
	if _, ok := storages[chatId]; !ok {
		storage = &inMemoryChatStorage{make([]*pb.Message, 0), make(chan chatStorageAction, 100), false}
		storages[chatId] = storage
	}
	return storage
}

type inMemoryChatStorage struct {
	messages []*pb.Message
	requests chan chatStorageAction
	acting   bool
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

func (s *inMemoryChatStorage) processAction(action chatStorageAction) {
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
}

func (s *inMemoryChatStorage) Act(cancel chan bool) {
	if s.acting {
		log.Panic("double act is forbidden")
	}
	s.acting = true
act:
	for {
		select {
		case <-cancel:
			break act
		case req := <-s.requests:
			s.processAction(req)
		}
	}
}
