package chatmanager

import (
	pb "whcrc/chat/proto"
)

type ChatId struct {
	SenderId   string
	ReceiverId string
}

type ReaderCounters struct {
	bufferOverflow uint64
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

// ChatReader is NOT thread-safe, concurrent execution of methods is not supported
type ChatReader interface {
	// Receives messages with ids in specified range: (fromId - count, fromId]
	LoadOldMessages(fromId uint64, count uint64) []*pb.Message

	// Blocks until new message is received. Wating could be cancelled via sending message to the `cancel` channel.
	Recv(cancel chan bool) (*pb.Message, error)

	// Method should be called once finished working with ChatReader entity. After Close() is called, usage of any methods on this instance
	// causes program to panic.
	Close()

	GetCounters() ReaderCounters
}

type ChatWriter interface {
	// Blocks until message is persisted and became visiable to readers.
	Send(msg *pb.Message) error

	// Method should be called once finished working with ChatWriter entity. After Close() is called, usage of any methods on this instance
	// causes program to panic.
	Close()
}

type ReaderConfig struct {
	bufferSize uint64
}

type ChatManager interface {
	GetReaderFor(chatId ChatId, config ReaderConfig) ChatReader
	GetWriterFor(chatId ChatId) ChatWriter
	Act()
	Close()
}

func CreateChatManager() ChatManager {
	return &chatManager{
		chats:    make(map[ChatId]*chat),
		requests: make(chan chatManagerRequest),
		closing:  false,
		closed:   make(chan bool),
	}
}
