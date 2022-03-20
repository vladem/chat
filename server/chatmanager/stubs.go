package chatmanager

import (
	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/common"
)

type ReaderCounters struct {
	bufferOverflow uint64
}

// ChatReader is NOT thread-safe, concurrent execution of methods is not supported
type ChatReader interface {
	// Receives messages with ids in specified range: (fromId - count, fromId]
	LoadOldMessages(fromId uint64, count uint64) []*pb.Message

	// Blocks until new message is received. Wating could be cancelled via sending message to the `cancel` channel.
	Recv(cancel <-chan struct{}) (*pb.Message, error)

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
	BufferSize uint64
}

type ChatManager interface {
	GetReaderFor(chatId cm.ChatId, config ReaderConfig) ChatReader
	GetWriterFor(chatId cm.ChatId) ChatWriter
	Act()
	Close()
}

func CreateChatManager() ChatManager {
	return &chatManager{
		chats:    make(map[cm.ChatId]*chat),
		requests: make(chan chatManagerRequest),
		closing:  false,
		closed:   make(chan bool),
	}
}
