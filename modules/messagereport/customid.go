package messagereport

import (
	"strings"
)

type CustomId struct {
	Action        string
	ChannelId     string
	MessageId     string
	UserId        string
	BaseMessageId string
}

func (c *CustomId) ToString() string {
	parts := make([]string, 0)

	if c.Action != "" {
		parts = append(parts, "action:"+c.Action)
	}
	if c.ChannelId != "" {
		parts = append(parts, "channel:"+c.ChannelId)
	}
	if c.MessageId != "" {
		parts = append(parts, "message:"+c.MessageId)
	}
	if c.UserId != "" {
		parts = append(parts, "user:"+c.UserId)
	}
	if c.BaseMessageId != "" {
		parts = append(parts, "base:"+c.BaseMessageId)
	}

	return strings.Join(parts, "-")
}

func (c *CustomId) FromString(source string) {
	for _, v := range strings.Split(source, "-") {
		parts := strings.SplitN(v, ":", 2)
		key := parts[0]
		value := parts[1]

		switch key {
		case "action":
			{
				c.Action = value
				break
			}
		case "channel":
			{
				c.ChannelId = value
				break
			}
		case "message":
			{
				c.MessageId = value
				break
			}
		case "user":
			{
				c.UserId = value
				break
			}
		case "base":
			{
				c.BaseMessageId = value
			}
		}
	}
}

func (c *CustomId) Clone() *CustomId {
	return &CustomId{
		Action:        c.Action,
		ChannelId:     c.ChannelId,
		MessageId:     c.MessageId,
		UserId:        c.UserId,
		BaseMessageId: c.BaseMessageId,
	}
}
