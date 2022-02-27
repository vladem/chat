package chatmanager

import (
	"fmt"
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type chatManagerRequest interface {
	isChatManagerRequest()
}

type getReaderRequest struct {
	chatId   ChatId
	config   ReaderConfig
	response chan *chatReader
}

func (r getReaderRequest) isChatManagerRequest() {
}

type getWriterRequest struct {
	chatId   ChatId
	response chan *chatWriter
}

func (r getWriterRequest) isChatManagerRequest() {
}

type chatStopRequest struct {
	chatId ChatId
}

func (r chatStopRequest) isChatManagerRequest() {
}

type chatManager struct {
	chats    map[ChatId]*chat
	requests chan chatManagerRequest
	closing  bool
	closed   chan bool
}

func (cm *chatManager) Close() {

}

func (cm *chatManager) GetReaderFor(chatId ChatId, config ReaderConfig) ChatReader {
	resp := make(chan *chatReader)
	cm.requests <- getReaderRequest{chatId: chatId, config: config, response: resp}
	return <-resp
}

func (cm *chatManager) GetWriterFor(chatId ChatId) ChatWriter {
	resp := make(chan *chatWriter)
	cm.requests <- getWriterRequest{chatId: chatId, response: resp}
	return <-resp
}

// private

func (cm *chatManager) processRequest(req chatManagerRequest) (proceed bool) {
	if req == nil {
		fmt.Printf("waiting for all chats to finish")
		cm.closing = true
		return true
	}

	switch request := req.(type) {
	case getReaderRequest:
		if cm.closing {
			log.Panic("[chat manager] request reader after close")
		}
		fmt.Printf("[chat manager] reader request for chat with id [%v]\n", request.chatId) // todo: logging without boilerplate
		request.response <- cm.getReader(request.chatId, request.config)
	case getWriterRequest:
		if cm.closing {
			log.Panic("[chat manager] request writer after close")
		}
		fmt.Printf("[chat manager] writer request for chat with id [%v]\n", request.chatId)
		request.response <- cm.getWriter(request.chatId)
	case chatStopRequest:
		chat, ok := cm.chats[request.chatId]
		if !ok {
			log.Fatalf("chat [%v] requested stop, but it's missing", request.chatId)
		}
		delete(cm.chats, request.chatId)
		chat.stopConfirmation <- true
		if len(cm.chats) == 0 {
			return false
		}
	}
	return true
}

func (cm *chatManager) Act() {
	fmt.Printf("[chat manager] started")
	for cm.processRequest(<-cm.requests) {
	}
	fmt.Printf("[chat manager] stopped")
}

func (cm *chatManager) getOrCreateChat(chatId ChatId) *chat {
	if chatDescr, ok := cm.chats[chatId]; ok {
		return chatDescr
	}
	chat := &chat{
		manager:           cm,
		chatId:            chatId,
		storage:           storage.GetChatStorage(chatId.String()), // todo: use ChatId struct
		broadcastRequests: make(chan *pb.Message),
		readers:           make(map[*chatReader]bool),
		readerRequests:    make(chan readerRequest),
		writersCount:      0,
		writerRequests:    make(chan writerRequest),
		active:            false,
		stopConfirmation:  make(chan bool),
	}
	cm.chats[chatId] = chat
	chat.start()
	return chat
}

func (cm *chatManager) getReader(chatId ChatId, config ReaderConfig) *chatReader {
	chat := cm.getOrCreateChat(chatId)
	reader := &chatReader{
		chat:         chat,
		buffer:       make(chan *pb.Message, config.bufferSize),
		closed:       false,
		unregistered: make(chan bool),
		err:          nil,
		errMessageId: 0,
	}
	req := readerRequest{
		reader:   reader,
		register: true,
		done:     make(chan bool),
	}
	chat.readerRequests <- req
	<-req.done
	return reader
}

func (cm *chatManager) getWriter(chatId ChatId) *chatWriter {
	chat := cm.getOrCreateChat(chatId)
	writer := &chatWriter{
		chat:         chat,
		closed:       false,
		unregistered: make(chan bool),
	}
	req := writerRequest{
		writer:   writer,
		register: true,
		done:     make(chan bool),
	}
	chat.writerRequests <- req
	<-req.done
	return writer
}

func (cm *chatManager) requestStop(chatId ChatId) {
	cm.requests <- chatStopRequest{chatId: chatId}
}
