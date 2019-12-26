package hjt

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"io/ioutil"
	"net/http"
	"strings"
)

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, c *discordgo.Channel, cmd string, args []string) {
	if len(args) == 0 {
		return
	}

	pasteLink := args[0]

	content, err := readFromUrlAsLowercase(pasteLink)
	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Invalid URL")
		logger.Debug().Printf("Invalid url: %s\n%s", pasteLink, err)
		return
	}

	db, err := database.Get()
	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
		return
	}

	var values []string
	err = db.Table("hjt").Where("? LIKE CONCAT('%', LOWER(name) ,'%')", content).Pluck("value", &values).Error

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		return
	}

	if len(values) == 0 {
		_, err = ds.ChannelMessageSend(c.ID, "No matches found")
	} else {
		_, err = ds.ChannelMessageSend(c.ID, strings.Join(values, ", "))
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
