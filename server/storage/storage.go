package storage

import (
	pb "whcrc/chat/proto"
)

type ChatStorage interface {
	Read(messageId uint64) chan *pb.Message
	Write(message *pb.Message) chan error // sets id of a message
	Act(cancel chan bool)
}

func GetChatStorage(chatId string) ChatStorage {
	return getInMemoryChatStorage(chatId)
}
