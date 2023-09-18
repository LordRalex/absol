package mcping

import (
	"bytes"
	"encoding/base64"
	"github.com/bwmarrin/discordgo"
	"github.com/iverly/go-mcping/mcping"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

var appId string

func (*Module) Load(ds *discordgo.Session) {
	appId = env.Get("discord.app_id")

	var guilds []string

	maps := env.GetStringArray("mcping.guilds", ";")
	for _, v := range maps {
		if v == "" {
			continue
		}

		guilds = append(guilds, v)
	}

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, v := range guilds {
			logger.Out().Printf("Registering %s for guild %s\n", mcpingOperation.Name, v)
			_, err := s.ApplicationCommandCreate(appId, v, mcpingOperation)
			if err != nil {
				logger.Err().Printf("Cannot create slash command %q: %v", mcpingOperation.Name, err)
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
				if i.ApplicationCommandData().Name == mcpingOperation.Name {
					runCommand(s, i)
				}
			}
		}
	})
}

var mcpingOperation = &discordgo.ApplicationCommand{
	Name:        "mcping",
	Description: "Checks to see if a server is online",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{

		{
			Name:        "host",
			Description: "IP:Port or Host of server",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

func runCommand(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	err := ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}

	data := i.ApplicationCommandData().Options[0]
	ip := data.StringValue()

	connectionSlice := strings.Split(ip, ":")

	port := 25565
	if len(connectionSlice) == 2 {
		port, err = strconv.Atoi(connectionSlice[1])
		if err != nil {
			msg := "That's not a valid port."
			_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			return
		}
	}

	// set up pinger
	pinger := mcping.NewPinger()
	response, err := pinger.PingWithTimeout(connectionSlice[0], uint16(port), 5*time.Second)
	if err != nil {
		// if it takes more than five seconds to ping, then the server is probably down
		msg := "Connecting to the server failed."
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	// set up the embed
	var fields []*discordgo.MessageEmbedField
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Latency",
		Value:  strconv.Itoa(int(response.Latency)) + "ms",
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Player Count",
		Value:  strconv.Itoa(response.PlayerCount.Online) + "/" + strconv.Itoa(response.PlayerCount.Max) + " players",
		Inline: false,
	})
	motdRegex := regexp.MustCompile(`§.`)
	if response.Motd == "" {
		response.Motd = "No motd detected for some reason."
	}
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Message Of The Day",
		Value:  motdRegex.ReplaceAllString(response.Motd, ""),
		Inline: false,
	})
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Version",
		Value:  response.Version,
		Inline: false,
	})

	// add image to embed
	embed := &discordgo.MessageEmbed{
		Title: "Ping response from `" + ip + "`",
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://favicon.png",
		},
		Fields: fields,
	}

	// add server favicon to the message
	var files []*discordgo.File
	if response.Favicon != "" {
		files = append(files, &discordgo.File{
			Name:        "favicon.png",
			ContentType: "image/png",
			Reader:      base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(strings.Split(response.Favicon, ",")[1]))),
		})
	} else {
		files = nil
	}

	_, err = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
		Files:  files,
	})
	if err != nil {
		logger.Err().Printf("Failed to send message\n%s", err)
	}
}

func (*Module) Name() string {
	return "mcping"
}
