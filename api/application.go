package api

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
)

func RegisterSlashCommnand(ds *discordgo.Session, guilds []string, cmd *discordgo.ApplicationCommand, function ApplicationCommand) {
	appId := env.Get("discord.app_id")

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, v := range guilds {
			logger.Out().Printf("Registering %s for guild %s\n", cmd.Name, v)
			_, err := s.ApplicationCommandCreate(appId, v, cmd)
			if err != nil {
				logger.Err().Printf("Cannot create slash command %q: %v", cmd.Name, err)
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
				if i.ApplicationCommandData().Name == cmd.Name {
					function(s, i)
				}
			}
		}
	})
}

type ApplicationCommand func(s *discordgo.Session, i *discordgo.InteractionCreate)
