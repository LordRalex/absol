package messagereport

import "strings"

type InteractionId struct {
	Action    string
	ChannelId string
	MessageId string
	UserId    string
}

func (c *InteractionId) ToString() string {
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

	return strings.Join(parts, "-")
}

func (c *InteractionId) FromString(source string) {
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
		}
	}
}

func (i *InteractionId) Clone() *InteractionId {
	return &InteractionId{
		Action:    i.Action,
		ChannelId: i.ChannelId,
		MessageId: i.MessageId,
		UserId:    i.UserId,
	}
}
