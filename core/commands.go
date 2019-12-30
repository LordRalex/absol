package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"strings"
)

var commandPrefix string

func init() {
	api.RegisterCommand("modules", RunModuleCommand)
}

func EnableCommands(session *discordgo.Session) {
	commandPrefix = viper.GetString("prefix")

	if commandPrefix == "" {
		commandPrefix = "!?"
	}

	logger.Out().Printf("Adding commands")
	session.AddHandler(onMessageCommand)
}

func onMessageCommand(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	if !strings.HasPrefix(mc.Message.Content, commandPrefix) {
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

	msg := strings.TrimPrefix(mc.Message.Content, commandPrefix)

	parts := strings.Split(msg, " ")
	cmd := parts[0]
	args := parts[1:]

	commandExecutor := api.GetCommand(cmd)

	if commandExecutor != nil {
		commandExecutor(ds, mc, cmd, args)
	}
}

func RunModuleCommand(session *discordgo.Session, mc *discordgo.MessageCreate, cmd string, args []string) {
	modules := make([]string, 0)
	for k, _ := range loadedModules {
		modules = append(modules, k)
	}
	_, _ = session.ChannelMessageSend(mc.ChannelID, "Registered: "+strings.Join(modules, ", "))
}
