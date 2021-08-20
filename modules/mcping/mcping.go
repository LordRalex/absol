package mcping

import (
	"bytes"
	"encoding/base64"
	"github.com/bwmarrin/discordgo"
	"github.com/iverly/go-mcping/mcping"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/lordralex/absol/modules/factoids"
	"regexp"
	"strconv"
	"strings"
)

type Module struct {
	api.Module
}

func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("mcping", RunCommand)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsDirectMessages)
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, _ string, args []string) {
	if len(args) == 0 {
		return
	}

	connectionSlice := strings.Split(args[0], ":")

	port := 25565

	var err error // prevent shadowing

	if len(connectionSlice) == 2 {
		port, err = strconv.Atoi(connectionSlice[1])
		if err != nil {
			err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "That's not a valid port!")
			return
		}
	}

	// set up pinger
	pinger := mcping.NewPinger()
	response, err := pinger.PingWithTimeout(connectionSlice[0], uint16(port), 5)
	if err != nil {
		// if it takes more then fice seconds to ping, then the server is probably down
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Connecting to the server failed.")
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
		Title: "Ping response from `" + strings.Join(args, "") + "`",
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

	send := &discordgo.MessageSend{
		Embed: embed,
		Files: files,
	}

	_, err = ds.ChannelMessageSendComplex(mc.ChannelID, send)
	if err != nil {
		logger.Err().Printf("Failed to send message\n%s", err)
	}
}
