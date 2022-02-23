package chatmanager

import (
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type chatDescriptor struct {
	instance                 *chat
	inflightRegisterRequests uint64
}

type getReaderRequest struct {
	chatId   ChatId
	response chan *chatReader
}

type getWriterRequest struct {
	chatId   ChatId
	response chan *chatWriter
}

type chatManager struct {
	chats          map[ChatId]chatDescriptor
	readerRequests chan getReaderRequest
	writerRequests chan getWriterRequest
	registerDone   chan ChatId
	stopRequests   chan ChatId
}

// public
func (cm *chatManager) Act() {
	go func() {
		cm.act()
	}()
}

func (cm *chatManager) GetReaderFor(chatId ChatId) ChatReader {
	resp := make(chan *chatReader)
	cm.readerRequests <- getReaderRequest{chatId: chatId, response: resp}
	return <-resp
}

func (cm *chatManager) GetWriterFor(chatId ChatId) ChatWriter {
	resp := make(chan *chatWriter)
	cm.writerRequests <- getWriterRequest{chatId: chatId, response: resp}
	return <-resp
}

// private

func (cm *chatManager) act() {
	for {
		select {
		case readerRequest := <-cm.readerRequests:
			readerRequest.response <- cm.getReader(readerRequest.chatId)
		case writerRequest := <-cm.writerRequests:
			writerRequest.response <- cm.getWriter(writerRequest.chatId)
		case chatId := <-cm.stopRequests:
			chatDescr, ok := cm.chats[chatId]
			if !ok {
				log.Fatalf("chat [%v] requested stop, but it's missing", chatId)
			}
			if chatDescr.inflightRegisterRequests != 0 {
				return
			}
			delete(cm.chats, chatId)
			chatDescr.instance.stopConfirmation <- true
		case chatId := <-cm.registerDone:
			if chatDescr, ok := cm.chats[chatId]; ok {
				chatDescr.inflightRegisterRequests--
			} else {
				log.Fatalf("chat [%v] said it's done register, but it's missing", chatId)
			}
		}
	}
}

func (cm *chatManager) getOrCreateChat(chatId ChatId) chatDescriptor {
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
	descr := chatDescriptor{instance: chat, inflightRegisterRequests: 0}
	cm.chats[chatId] = descr
	chat.start()
	return descr
}

func (cm *chatManager) getReader(chatId ChatId) *chatReader {
	chatDescr := cm.getOrCreateChat(chatId)
	reader := &chatReader{
		chat:         chatDescr.instance,
		buffer:       make(chan *pb.Message),
		closed:       false,
		unregistered: make(chan bool),
		err:          nil,
		errMessageId: 0,
	}
	chatDescr.instance.readerRequests <- readerRequest{
		reader:   reader,
		register: true,
	}
	chatDescr.inflightRegisterRequests++
	return reader
}

func (cm *chatManager) getWriter(chatId ChatId) *chatWriter {
	chatDescr := cm.getOrCreateChat(chatId)
	writer := &chatWriter{
		chat:         chatDescr.instance,
		closed:       false,
		unregistered: make(chan bool),
	}
	chatDescr.instance.writerRequests <- writerRequest{
		writer:   writer,
		register: true,
	}
	chatDescr.inflightRegisterRequests++
	return writer
}

func (cm *chatManager) requestStop(chatId ChatId) {
	cm.stopRequests <- chatId
}
