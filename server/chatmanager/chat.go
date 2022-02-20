package chatmanager

import (
	"fmt"
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type chat struct {
	manager             *chatManager
	chatId              ChatId
	storage             storage.ChatStorage
	notifications       chan *pb.Message
	outputChannels      map[OutputChannel]bool
	subscribeRequests   chan OutputChannel
	unsubscribeRequests chan unsubscribeRequest
	acting              bool
	stopConfirmation    chan bool
}

// public
func (c *chat) LoadOldMessages(fromId uint64, toId uint64) OutputChannel {
	out := c.storage.Read(toId) // todo: load more than one message
	return out
}

func (c *chat) Communicate(cancel chan bool) (in InputChannel, out OutputChannel) {
	in = make(InputChannel)
	out = c.subscribeForNewMessages()
	go c.processIncommingMessages(in, out, cancel)
	return
}

// private
type unsubscribeRequest struct {
	outputChannel OutputChannel
	done          chan bool
}

func (c *chat) broadcast(message *pb.Message) {
	for output := range c.outputChannels {
		output <- message
	}
}

func (c *chat) act() {
	if c.acting {
		log.Panic("double act is forbidden")
	}
	c.acting = true
act:
	for {
		select {
		case outputChannel := <-c.subscribeRequests:
			c.outputChannels[outputChannel] = true
		case unsubscribeRequest := <-c.unsubscribeRequests:
			delete(c.outputChannels, unsubscribeRequest.outputChannel)
			unsubscribeRequest.done <- true
			if len(c.outputChannels) == 0 {
				c.manager.requestStop(c.chatId)
			}
		case message := <-c.notifications:
			c.broadcast(message)
		case <-c.stopConfirmation:
			break act
		}
	}
	fmt.Printf("chat [%v] stopped", c.chatId)
}

func (c *chat) subscribeForNewMessages() OutputChannel {
	outputChannel := make(OutputChannel)
	c.subscribeRequests <- outputChannel
	return outputChannel
}

func (c *chat) unsubscribe(outputChannel OutputChannel) {
	req := unsubscribeRequest{
		outputChannel: outputChannel,
		done:          make(chan bool),
	}
	c.unsubscribeRequests <- req
	<-req.done
}

func (c *chat) processIncommingMessages(in InputChannel, out OutputChannel, cancel chan bool) {
	defer c.unsubscribe(out)
input:
	for {
		select {
		case <-cancel:
			break input
		case message := <-in:
			errChan := c.storage.Write(message)
			err := <-errChan
			if err != nil {
				fmt.Printf("failed to write message [%s] to chat [%v] with error [%v]\n", string(message.Data), c.chatId, err)
			} else {
				c.notifications <- message
			}
		}
	}
}
