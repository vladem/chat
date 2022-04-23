package chatmanager

import (
	"log"
	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/common"
	"whcrc/chat/server/storage"
)

type chatManagerRequest interface {
	isChatManagerRequest()
}

type getReaderRequest struct {
	chatId   cm.ChatId
	config   ReaderConfig
	response chan *chatReader
}

func (r getReaderRequest) isChatManagerRequest() {
}

type getWriterRequest struct {
	chatId   cm.ChatId
	response chan *chatWriter
}

func (r getWriterRequest) isChatManagerRequest() {
}

type chatStopRequest struct {
	chatId cm.ChatId
}

func (r chatStopRequest) isChatManagerRequest() {
}

type chatManager struct {
	chats    map[cm.ChatId]*chat
	requests chan chatManagerRequest
	closing  bool
	closed   chan bool
}

func (cm *chatManager) Close() {

}

func (cm *chatManager) GetReaderFor(chatId cm.ChatId, config ReaderConfig) ChatReader {
	resp := make(chan *chatReader)
	cm.requests <- getReaderRequest{chatId: chatId, config: config, response: resp}
	return <-resp
}

func (cm *chatManager) GetWriterFor(chatId cm.ChatId) ChatWriter {
	resp := make(chan *chatWriter)
	cm.requests <- getWriterRequest{chatId: chatId, response: resp}
	return <-resp
}

// private

func (cm *chatManager) processRequest(req chatManagerRequest) (proceed bool) {
	if req == nil {
		log.Printf("waiting for all chats to finish\n")
		cm.closing = true
		return true
	}

	switch request := req.(type) {
	case getReaderRequest:
		if cm.closing {
			log.Panic("[chat manager] request reader after close")
		}
		log.Printf("[chat manager] reader request for chat with id [%v]\n", request.chatId)
		request.response <- cm.getReader(request.chatId, request.config)
	case getWriterRequest:
		if cm.closing {
			log.Panic("[chat manager] request writer after close")
		}
		log.Printf("[chat manager] writer request for chat with id [%v]\n", request.chatId)
		request.response <- cm.getWriter(request.chatId)
	case chatStopRequest:
		chat, ok := cm.chats[request.chatId]
		if !ok {
			log.Fatalf("chat [%v] requested stop, but it's missing", request.chatId)
		}
		delete(cm.chats, request.chatId)
		chat.stopConfirmation <- true
		if len(cm.chats) == 0 && cm.closing {
			return false
		}
	}
	return true
}

func (cm *chatManager) Act() {
	log.Printf("[chat manager] started\n")
	for cm.processRequest(<-cm.requests) {
	}
	log.Printf("[chat manager] stopped\n")
}

func (cm *chatManager) getOrCreateChat(chatId cm.ChatId) *chat {
	if chatDescr, ok := cm.chats[chatId]; ok {
		return chatDescr
	}
	chat := &chat{
		manager:                 cm,
		chatId:                  chatId,
		storage:                 storage.GetChatStorage(chatId),
		broadcastRequests:       make(chan *pb.Message),
		lastBroadcastedId:       0,
		readers:                 make(map[*chatReader]bool),
		readerRequests:          make(chan readerRequest),
		writersCount:            0,
		writerRequests:          make(chan writerRequest),
		active:                  false,
		stopConfirmation:        make(chan bool),
		suspendedReaderRequests: make(chan suspendedReaderRequest),
	}
	cm.chats[chatId] = chat
	chat.start()
	return chat
}

func (cm *chatManager) getReader(chatId cm.ChatId, config ReaderConfig) *chatReader {
	chat := cm.getOrCreateChat(chatId)
	reader := &chatReader{
		chat:         chat,
		buffer:       make(chan *pb.Message, config.BufferSize),
		closed:       false,
		unregistered: make(chan bool, 1),
		suspend:      make(chan bool, 1),
		suspended:    false,
		counters:     ReaderCounters{},
		lastReadId:   0,
	}
	req := readerRequest{
		reader:   reader,
		register: true,
		done:     make(chan bool, 1),
	}
	chat.readerRequests <- req
	<-req.done
	return reader
}

func (cm *chatManager) getWriter(chatId cm.ChatId) *chatWriter {
	chat := cm.getOrCreateChat(chatId)
	writer := &chatWriter{
		chat:         chat,
		closed:       false,
		unregistered: make(chan bool, 1),
	}
	req := writerRequest{
		writer:   writer,
		register: true,
		done:     make(chan bool, 1),
	}
	chat.writerRequests <- req
	<-req.done
	return writer
}

func (cm *chatManager) requestStop(chatId cm.ChatId) {
	cm.requests <- chatStopRequest{chatId: chatId}
}
