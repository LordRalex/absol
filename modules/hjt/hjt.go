package hjt

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"regexp"
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

	content, err := readFromUrl(pasteLink)
	if err != nil {
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "Invalid URL")
		return
	}

	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Failed to connect to database\n%s", err)
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	var values []HJT
	err = db.Table("hjts").Select([]string{"id", "name", "match_criteria"}).Find(&values).Error

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	var results []uint
	for _, v := range values {
		matches, err := regexp.Match(v.MatchCriteria, content)
		if err != nil {
			_, _ = ds.ChannelMessageSend(mc.ChannelID, v.Name + " has a bad regex statement: " + err.Error())
			return
		}
		if matches {
			results = append(results, v.Id)
		}
	}

	if len(results) == 0 {
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "No matches found for HJT")
		return
	}

	var data []HJT
	err = db.Find(&data, results).Order("severity desc").Error
	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "Failed to connect to database")
		return
	}

	message := ""
	for i, v := range data {
		if i != 0 {
			message += "\n"
		}
		message += v.SeverityEmoji + " [" + v.Category + "] " + v.Name + ": " + v.Description
	}
	_, _ = ds.ChannelMessageSend(mc.ChannelID, message)
}

func readFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

type HJT struct {
	Id uint
	Name string
	MatchCriteria string
	Description string
	Category string
	Severity Severity
	SeverityEmoji string `gorm:"-"`
}

type Severity int

var SeverityInfo Severity = 0
var SeverityLow Severity = 1
var SeverityMedium Severity = 2
var SeverityHigh Severity = 3

func (s Severity) ToEmojiString() string {
	switch s {
	case SeverityHigh:
		return ":red_circle:"
	case SeverityLow:
		return ":yellow_circle:"
	case SeverityMedium:
		return ":orange_circle:"
	default:
		return ":green_circle:"
	}
}

func (h *HJT) AfterFind(tx *gorm.DB) (err error) {
	h.SeverityEmoji = h.Severity.ToEmojiString()
	if h.Name == "" {
		h.Name = h.MatchCriteria
	}
	return
}