package common

type ChatId struct {
	first  string
	second string
}

func GetChatId(senderId string, receiverId string) ChatId {
	if senderId > receiverId {
		return ChatId{
			first:  receiverId,
			second: senderId,
		}
	}
	return ChatId{
		first:  senderId,
		second: receiverId,
	}
}

func (c ChatId) String() string {
	return c.first + "<->" + c.second
}
