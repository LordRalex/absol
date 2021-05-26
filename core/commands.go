package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
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

	session.AddHandler(onMessageCommand)
}

func onMessageCommand(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	if !strings.HasPrefix(mc.Message.Content, commandPrefix) {
		return
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
