package chatmanager

import (
	"errors"
	"fmt"
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

// Created and started by chatManager. Stopped once all readers and writers were closed and chat is unregistered from chatManager.
// At most one active instance of this struct is present for each ChatId within a process. As it been said, chat is an 'active' instance,
// meaning it has a running goroutine associated with it, which serves as a multiplexor of incomming messages. This struct also holds some
// resources which should be shared between all readers and writers of single chat (currently a storage, which requires a program-scope lock to be held
// for instantiation).
type chat struct {
	manager *chatManager
	chatId  ChatId
	storage storage.ChatStorage

	broadcastRequests chan *pb.Message

	readers        map[*chatReader]bool
	readerRequests chan readerRequest

	writersCount   uint64
	writerRequests chan writerRequest

	active           bool
	stopConfirmation chan bool
}

type readerRequest struct {
	reader   *chatReader
	register bool
}

type writerRequest struct {
	writer   *chatWriter
	register bool
}

// Implements ChatReader interface. Created by chatManager.
type chatReader struct {
	chat         *chat
	buffer       chan *pb.Message
	closed       bool
	unregistered chan bool
	err          error
	errMessageId uint64
}

// Implements ChatWriter interface. Created by chatManager.
type chatWriter struct {
	chat         *chat
	closed       bool
	unregistered chan bool
}

func (c *chat) broadcast(message *pb.Message) {
	readersToUnregister := make(map[*chatReader]bool)
	for reader := range c.readers {
		select {
		case reader.buffer <- message:
			fmt.Printf("sent message [%s] to reader [%p]\n", message.Data, reader)
		default:
			fmt.Printf("reader's [%p] buffer is full, closing it\n", reader)
			reader.err = errors.New("buffer is full")
			reader.errMessageId = message.MessageId
			readersToUnregister[reader] = true
		}
	}

	for reader := range readersToUnregister {
		delete(c.readers, reader)
		reader.unregistered <- true
	}
}

func (c *chat) processOneRequest() (proceed bool) {
	select {
	case req := <-c.readerRequests:
		if req.register {
			c.readers[req.reader] = true
			c.manager.registerDone <- c.chatId
		} else {
			delete(c.readers, req.reader)
			req.reader.unregistered <- true
		}
	case req := <-c.writerRequests:
		if req.register {
			c.writersCount++
			c.manager.registerDone <- c.chatId
		} else {
			c.writersCount--
			req.writer.unregistered <- true
		}
	case message := <-c.broadcastRequests:
		c.broadcast(message)
	case <-c.stopConfirmation:
		return false
	}

	if len(c.readers) == 0 && c.writersCount == 0 {
		c.manager.requestStop(c.chatId)
	}

	return true
}

func (c *chat) start() {
	go func() {
		if c.active {
			log.Panicf("chat with id [%v] is already started", c.chatId)
		}
		storageCancel := make(chan bool)
		go c.storage.Act(storageCancel)
		c.active = true
		fmt.Printf("chat with id [%v] started\n", c.chatId)
		for c.processOneRequest() {
		}
		storageCancel <- true
		c.active = false
		fmt.Printf("chat with id [%v] stopped\n", c.chatId)
	}()
}

func (r *chatReader) LoadOldMessages(fromId uint64, count uint64) chan *pb.Message {
	// todo: implement me
	res := make(chan *pb.Message, 1)
	res <- nil
	return res
}

func (r *chatReader) Recv(cancel chan bool) (*pb.Message, error) {
	if r.closed {
		panic("using closed reader")
	}
	select {
	case message := <-r.buffer:
		return message, nil
	case <-r.unregistered:
		r.closed = true
		return nil, r.err
	case <-cancel:
		return nil, nil
	}
}

func (r *chatReader) Close() {
	if r.closed {
		panic("using closed reader")
	}
	r.chat.readerRequests <- readerRequest{
		reader:   r,
		register: false,
	}
	<-r.unregistered
	r.closed = true
}

func (w *chatWriter) Send(msg *pb.Message) error {
	if w.closed {
		panic("using closed writer")
	}
	errChan := w.chat.storage.Write(msg)
	err := <-errChan
	if err != nil {
		fmt.Printf("failed to write message [%s] to chat [%v] with error [%v]\n", string(msg.Data), w.chat.chatId, err)
		return err
	}
	w.chat.broadcastRequests <- msg
	return nil
}

func (w *chatWriter) Close() {
	if w.closed {
		panic("using closed writer")
	}
	w.chat.writerRequests <- writerRequest{
		writer:   w,
		register: false,
	}
	<-w.unregistered
	w.closed = true
}
