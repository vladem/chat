package chatmanager

import (
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type chatManager struct {
	chats        map[ChatId]*chat
	getRequests  chan getRequest
	stopRequests chan ChatId
}

// public
func (cm *chatManager) Act() {
	for {
		select {
		case getRequest := <-cm.getRequests:
			getRequest.response <- cm.getChat(getRequest.chatId)
		case chatId := <-cm.stopRequests:
			chat, ok := cm.chats[chatId]
			if !ok {
				log.Fatalf("chat %v requested stop, but it's missing", chatId)
			}
			delete(cm.chats, chatId)
			chat.stopConfirmation <- true
		}
	}
}

func (cm *chatManager) GetChat(chatId ChatId) chan Chat {
	resp := make(chan Chat)
	cm.getRequests <- getRequest{chatId: chatId, response: resp}
	return resp
}

// private
type getRequest struct {
	chatId   ChatId
	response chan Chat
}

func (cm *chatManager) getChat(chatId ChatId) Chat {
	if chat, ok := cm.chats[chatId]; ok {
		return chat
	}
	chat := &chat{
		manager:             cm,
		chatId:              chatId,
		storage:             storage.GetChatStorage(chatId.String()), // todo: use chatId in storage interface
		notifications:       make(chan *pb.Message),
		outputChannels:      make(map[OutputChannel]bool),
		subscribeRequests:   make(chan OutputChannel),
		unsubscribeRequests: make(chan unsubscribeRequest),
		acting:              false,
		stopConfirmation:    make(chan bool),
	}
	cm.chats[chatId] = chat
	go chat.act()
	return chat
}

func (cm *chatManager) requestStop(chatId ChatId) {
	cm.stopRequests <- chatId
}
