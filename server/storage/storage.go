package storage

import (
	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/common"
)

type ChatStorage interface {
	Read(messageId uint64) chan *pb.Message
	Write(message *pb.Message) chan error // sets id of a message
	Act()
	Close()
}

func GetChatStorage(chatId cm.ChatId) ChatStorage {
	return getInMemoryChatStorage(chatId)
}
