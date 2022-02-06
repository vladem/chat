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
	LoadOldMessages(fromId uint64, toId uint64) chan *pb.Message
	Communicate(cancel chan bool) (in InputChannel, out OutputChannel)
	Act()
}

type ChatManager interface {
	GetChat(chatId ChatId) chan Chat
}

func GetChatManager(chatId string) ChatManager {
	// return &chatManager{storage.GetChatStorage(chatId), make(chan *pb.Message, 100)}
	return nil
}
