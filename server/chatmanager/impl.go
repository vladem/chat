package chatmanager

import (
	"fmt"
	"log"
	pb "whcrc/chat/proto"
	"whcrc/chat/server/storage"
)

type unsubscribeRequest struct {
	outputChannel OutputChannel
	done          chan bool
}

type chat struct {
	chatId              ChatId
	storage             storage.ChatStorage
	notifications       chan *pb.Message
	outputChannels      map[OutputChannel]bool
	subscribeRequests   chan OutputChannel
	unsubscribeRequests chan unsubscribeRequest
	acting              bool
}

func (c *chat) broadcast(message *pb.Message) {
	for output := range c.outputChannels {
		output <- message
	}
}

func (c *chat) Act() {
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
				break act
			}
		case message := <-c.notifications:
			c.broadcast(message)
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

func (c *chat) processIncommingMessages(in InputChannel, cancel chan bool) {
input:
	for {
		select {
		case <-cancel:
			break input
		case message := <-in:
			errChan := c.storage.Write(message)
			err := <-errChan
			if err != nil {
				fmt.Errorf("failed to write message [%s] to chat [%v] with error [%v]\n", string(message.Data), c.chatId, err)
			} else {
				c.notifications <- message
			}
		}
	}
}

func (c *chat) Communicate(cancel chan bool) (in InputChannel, out OutputChannel) {
	in = make(InputChannel)
	out = c.subscribeForNewMessages()
	defer c.unsubscribe(out)

	inputRoutineCancel := make(chan bool)
	go c.processIncommingMessages(in, inputRoutineCancel)
	<-cancel
	inputRoutineCancel <- true
	return in, out
}

// !!! stopped here, following actions:
// - implement chat manager (where chats are created if needed)
// - implement load of old messages for chat
// - fix proto, so it's a bidirectional stream
// - use chat manager / chat in handlers

// type chats struct {
// 	chats map[ChatId]*chatManager
// }

// func (c chats) get(senderId, receiverId string) *chatManager {
// 	chatId := c.getChatId(senderId, receiverId)
// 	if chat, ok := c.chats[chatId]; !ok {
// 		c.chats[chatId] = NewChatManager(chatId)
// 	}
// 	return chat
// }
