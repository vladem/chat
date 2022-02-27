package chatmanager

import (
	pb "whcrc/chat/proto"
)

type ChatId struct {
	SenderId   string
	ReceiverId string
}

func (c ChatId) String() string {
	// todo: name collisions?
	var chatId string
	if c.SenderId > c.ReceiverId {
		chatId = c.ReceiverId + c.SenderId
	} else {
		chatId = c.SenderId + c.ReceiverId
	}
	return chatId
}

type ChatReader interface {
	// Receives messages with ids in specified range: (fromId - count, fromId]. Nil-value received from the channel specifies end of the stream.
	LoadOldMessages(fromId uint64, count uint64) chan *pb.Message

	// Blocks until new message is received. Wating could be cancelled via sending message to the `cancel` channel.
	Recv(cancel chan bool) (*pb.Message, error)

	// Method should be called once finished working with ChatReader entity. After Close() is called, usage of any methods on this instance
	// causes program to panic.
	Close()
}

type ChatWriter interface {
	// Blocks until message is persisted and became visiable to readers.
	Send(msg *pb.Message) error

	// Method should be called once finished working with ChatWriter entity. After Close() is called, usage of any methods on this instance
	// causes program to panic.
	Close()
}

type ChatManager interface {
	GetReaderFor(chatId ChatId) ChatReader
	GetWriterFor(chatId ChatId) ChatWriter
	Act()
}

func CreateChatManager() ChatManager {
	return &chatManager{
		chats:          make(map[ChatId]chatDescriptor),
		readerRequests: make(chan getReaderRequest),
		writerRequests: make(chan getWriterRequest),
		registerDone:   make(chan ChatId),
		stopRequests:   make(chan ChatId),
	}
}
