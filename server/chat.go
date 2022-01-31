package main

import (
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type chatManager struct {
	storage       storage.ChatStorage
	notifications chan *pb.Message
}

func NewChatManager(chatId string) *chatManager {
	return &chatManager{storage.GetChatStorage(chatId), make(chan *pb.Message, 100)}
}

type chats struct {
	chats map[string]*chatManager
}

// func (c chats) Get(senderId, receiverId string) *chatManager {
// 	var chatId string
// 	if senderId > receiverId {
// 		chatId = receiverId + senderId
// 	} else {
// 		chatId = senderId + receiverId
// 	}

// 	if chat, ok := c.chats[chatId]; !ok {
// 		c.chats[chatId] = NewChatManager(chatId)
// 	}
// 	return chat
// }

type ReaderWriter interface {
	Act(done chan bool)
}
