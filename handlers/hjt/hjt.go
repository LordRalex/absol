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

	var data []hjt
	err = db.Find(&data).Error

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		return
	}

	var result strings.Builder

	for _, value := range data {
		if strings.Contains(content, strings.ToLower(value.Name)) {
			if result.Len() != 0 {
				result.WriteString(", ")
			}
			result.WriteString(value.Value)
		}
	}

	if result.Len() == 0 {
		_, err = ds.ChannelMessageSend(c.ID, "No matches found")
	} else {
		_, err = ds.ChannelMessageSend(c.ID, result.String())
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

type hjt struct {
	Name  string `gorm:"name"`
	Value string `gorm:"value"`
}

func (hjt) TableName() string {
	return "hjt"
}
