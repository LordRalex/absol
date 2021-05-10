package hjt

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/gorm"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Module struct {
	api.Module
}

func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("hjt", RunCommand)
	api.RegisterCommand("hjtf", RunCommandFull)

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

	matches, err := getMatches(content, db)

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	if len(matches) == 0 {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "No matches found")
	} else {
		_, err = ds.ChannelMessageSend(mc.ChannelID, strings.Join(matches, ", "))
	}
}

func RunCommandFull(ds *discordgo.Session, mc *discordgo.MessageCreate, cmd string, args []string) {
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

	matches, err := getMatches(content, db)

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	var objs []HJT
	err = db.Table("hjt").Find(&objs, matches).Error
	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
		return
	}

	if len(objs) == 0 {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "No matches found")
	} else {
		var sb strings.Builder
		sb.WriteString("Make sure to always double check the log manually before recommending any action\n")
		for _, obj := range objs {
			sb.WriteString(obj.severity + " ")
			sb.WriteString(obj.category + " ")
			sb.WriteString(obj.name + " ")
			sb.WriteString(obj.value)
			sb.WriteString("\n")
		}
		_, err = ds.ChannelMessageSend(mc.ChannelID, sb.String())
	}
}

func getMatches(content string, db *gorm.DB) ([]string, error) {
	var keys []string
	err := db.Table("hjt").Pluck("name", &keys).Error
	if err != nil {
		return nil, err
	}

	var matches []string

	for _, v := range keys {
		matched, _ := regexp.MatchString(v, content)
		if matched {
			matches = append(matches, v)
		}
	}

	return matches, nil
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

type HJT struct {
	name     string `gorm:"primaryKey"`
	value    string
	category string
	severity string
}
