package chatmanager

import (
	pb "whcrc/chat/proto"
)

type ChatId struct {
	SenderId   string
	ReceiverId string
}

func (c ChatId) String() string {
	var chatId string
	if c.SenderId > c.ReceiverId {
		chatId = c.ReceiverId + c.SenderId
	} else {
		chatId = c.SenderId + c.ReceiverId
	}
	return chatId
}

type InputChannel = chan *pb.Message
type OutputChannel = chan *pb.Message

type Chat interface {
	LoadOldMessages(fromId uint64, toId uint64) OutputChannel
	Communicate(cancel chan bool) (in InputChannel, out OutputChannel)
}

type ChatManager interface {
	GetChat(chatId ChatId) chan Chat
	Act()
}

func CreateChatManager() ChatManager {
	return &chatManager{
		chats:        make(map[ChatId]*chat),
		getRequests:  make(chan getRequest),
		stopRequests: make(chan ChatId),
	}
}
