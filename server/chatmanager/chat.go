package chatmanager

import (
	"errors"
	"log"
	pb "whcrc/chat/proto"
	cm "whcrc/chat/server/common"
	"whcrc/chat/server/storage"
)

// Created and started by chatManager. Stopped once all readers and writers were closed and chat is unregistered from chatManager.
// At most one active instance of this struct is present for each ChatId within a process. As it been said, chat is an 'active' instance,
// meaning it has a running goroutine associated with it, which serves as a multiplexor of incomming messages. This struct also holds some
// resources which should be shared between all readers and writers of single chat (currently a storage, which requires a program-scope lock to be held
// for instantiation).
type chat struct {
	manager *chatManager
	chatId  cm.ChatId
	storage storage.ChatStorage

	broadcastRequests chan *pb.Message
	lastBroadcastedId uint64

	readers        map[*chatReader]bool // reader -> isActive
	readerRequests chan readerRequest

	writersCount   uint64
	writerRequests chan writerRequest

	active           bool
	stopConfirmation chan bool

	suspendedReaderRequests chan suspendedReaderRequest
}

type suspenedReaderResponse struct {
	message     *pb.Message
	activeAgain bool
}

type suspendedReaderRequest struct {
	reader    *chatReader
	messageId uint64
	response  chan suspenedReaderResponse
}

type readerRequest struct {
	reader   *chatReader
	register bool
	done     chan bool
}

type writerRequest struct {
	writer   *chatWriter
	register bool
	done     chan bool
}

// Implements ChatReader interface. Created by chatManager.
type chatReader struct {
	chat         *chat
	buffer       chan *pb.Message
	closed       bool
	unregistered chan bool
	suspend      chan bool
	suspended    bool
	counters     ReaderCounters
	lastReadId   uint64
}

// Implements ChatWriter interface. Created by chatManager.
type chatWriter struct {
	chat         *chat
	closed       bool
	unregistered chan bool
}

func (c *chat) broadcast(message *pb.Message) {
	for reader, isActive := range c.readers {
		if !isActive {
			continue
		}
		select {
		case reader.buffer <- message:
			log.Printf("sent message [%s] to reader [%p]\n", message.Data, reader)
		default:
			log.Printf("reader's [%p] buffer is full, suspending it\n", reader)
			raiseFlag(reader.suspend)
			c.readers[reader] = false // suspend reader
		}
	}
}

func raiseFlag(flag chan bool) {
	select {
	case flag <- true:
		break
	default:
		panic("blocking flag")
	}
}

func (c *chat) processOneRequest() (proceed bool) {
	select {
	case req := <-c.readerRequests:
		if req.register {
			c.readers[req.reader] = true
			raiseFlag(req.done)
		} else {
			delete(c.readers, req.reader)
			raiseFlag(req.reader.unregistered)
		}
	case req := <-c.writerRequests:
		if req.register {
			c.writersCount++
			raiseFlag(req.done)
		} else {
			c.writersCount--
			raiseFlag(req.writer.unregistered)
		}
	case message := <-c.broadcastRequests:
		c.broadcast(message)
		c.lastBroadcastedId = message.MessageId
	case <-c.stopConfirmation:
		return false
	case req := <-c.suspendedReaderRequests:
		resp := suspenedReaderResponse{}
		resp.message = <-c.storage.Read(req.messageId) // todo: seems like inefficient to wait here
		if resp.message.MessageId == c.lastBroadcastedId {
			c.readers[req.reader] = true
			resp.activeAgain = true
		}
		req.response <- resp // nonblocking
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
		go c.storage.Act()
		c.active = true
		log.Printf("chat with id [%v] started\n", c.chatId)
		for c.processOneRequest() {
		}
		c.storage.Close()
		c.active = false
		log.Printf("chat with id [%v] stopped\n", c.chatId)
	}()
}

func (r *chatReader) LoadOldMessages(fromId uint64, count uint64) []*pb.Message {
	// todo: implement me
	return []*pb.Message{}
}

func (r *chatReader) suspenedRecv() *pb.Message {
	// read from buffer, until it's empty
	select {
	case message := <-r.buffer:
		r.lastReadId = message.MessageId
		return message
	default:
		break
	}

	// receive by communicating directly with chat instance (instead of "listening" to buffer of broadcasted messages)
	req := suspendedReaderRequest{
		reader:    r,
		messageId: r.lastReadId + 1,
		response:  make(chan suspenedReaderResponse, 1),
	}
	r.chat.suspendedReaderRequests <- req
	resp := <-req.response
	if resp.activeAgain {
		r.suspended = false
	}
	r.lastReadId += 1
	return resp.message
}

func (r *chatReader) Recv(cancel <-chan struct{}) (*pb.Message, error) {
	if r.closed {
		panic("using closed reader")
	}
	if r.suspended {
		return r.suspenedRecv(), nil
	}
	select {
	case message := <-r.buffer:
		r.lastReadId = message.MessageId
		return message, nil
	case <-r.unregistered:
		r.closed = true
		return nil, errors.New("unregistered")
	case <-r.suspend:
		r.counters.bufferOverflow++
		r.suspended = true
		return r.suspenedRecv(), nil
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

func (r *chatReader) GetCounters() ReaderCounters {
	return r.counters
}

func (w *chatWriter) Send(msg *pb.Message) error {
	if w.closed {
		panic("using closed writer")
	}
	errChan := w.chat.storage.Write(msg)
	err := <-errChan
	if err != nil {
		log.Printf("failed to write message [%s] to chat [%v] with error [%v]\n", string(msg.Data), w.chat.chatId, err)
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
