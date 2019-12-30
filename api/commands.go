package api

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

var registeredCommands = make(map[string]CommandFunc)

func RegisterCommand(cmd string, commandFunc CommandFunc) {
	registeredCommands[strings.ToLower(cmd)] = commandFunc
}

func GetCommand(cmd string) CommandFunc {
	executor := registeredCommands[strings.ToLower(cmd)]
	if executor == nil {
		return registeredCommands[""]
	}
	return executor
}

type CommandFunc func(session *discordgo.Session, message *discordgo.MessageCreate, cmd string, args []string)
