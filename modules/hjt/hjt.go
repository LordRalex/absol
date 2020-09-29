package hjt

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/lordralex/absol/api/database"
	"io/ioutil"
	"net/http"
	"strings"
)

type Module struct {
	api.Module
}

func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("hjt", RunCommand)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsDirectMessages)
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, cmd string, args []string) {
	if len(args) == 0 {
		return
	}
	pasteLink := args[0]

	content, err := readFromUrlAsLowercase(pasteLink)
	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Invalid URL")
		return
	}

	db, err := database.Get()
	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
		return
	}

	var values []string
	err = db.Table("hjt").Where("? LIKE CONCAT('%', LOWER(name) ,'%')", content).Pluck("value", &values).Error

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	if len(values) == 0 {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "No matches found")
	} else {
		_, err = ds.ChannelMessageSend(mc.ChannelID, strings.Join(values, ", "))
	}
}

func readFromUrlAsLowercase(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}
	return strings.ToLower(string(data)), nil
}
