package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/handlers/twitch"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"strings"
)

var CommandPrefix string

func RegisterCommands(session *discordgo.Session) {
	CommandPrefix = viper.GetString("prefix")

	if CommandPrefix == "" {
		CommandPrefix = "!?"
	}

	logger.Out().Printf("Adding commands")
	session.AddHandler(OnMessageCommand)
}

func OnMessageCommand(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	if !strings.HasPrefix(mc.Message.Content, CommandPrefix) {
		return
	}

	logger.Out().Printf("Command receieved")

	c, err := ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
		if err != nil {
			logger.Err().Printf("unable to fetch Channel for Message, %s", err)
		} else {
			// Attempt to add this channel into our State
			err = ds.State.ChannelAdd(c)
			if err != nil {
				logger.Err().Printf("error updating State with Channel, %s", err)
			}
		}
	}

	//only work with DMs
	if c == nil || c.Type != discordgo.ChannelTypeDM {
		return
	}

	msg := strings.TrimPrefix(mc.Message.Content, CommandPrefix)

	parts := strings.Split(msg, " ")
	cmd := parts[0]
	args := parts[1:]

	switch strings.ToLower(cmd) {
	case "twitchid":
		{
			twitch.RunCommand(ds, mc, c, cmd, args)
		}
	case "mutemcf":
		{
			twitch.RunCommand(ds, mc, c, cmd, args)
		}
	}
}
