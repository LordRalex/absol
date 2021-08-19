package search

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

// Load absol commands API
func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("s", RunCommand)
	api.RegisterCommand("search", RunCommand)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsDirectMessages)
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, _ string, args []string) {
	if len(args) == 0 {
		err := SendWithSelfDelete(ds, mc.ChannelID, "You must specify a search string!")
		if err != nil {
			return
		}
		return
	} else if len(strings.Join(args, "")) < 3 {
		err := SendWithSelfDelete(ds, mc.ChannelID, "Your search is too short!")
		if err != nil {
			return
		}
		return
	}

	max := viper.GetInt("factoids.max")
	if max == 0 {
		max = 5
	}

	db, err := database.Get()
	if err != nil {
		err = SendWithSelfDelete(ds, mc.ChannelID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
		return
	}

	// gets the factoids table
	var factoidsList []factoid
	err = db.Find(&factoidsList).Error
	if err != nil {
		return
	}

	pageNumber := 0
	pageNumber, err = strconv.Atoi(args[len(args)-1]) // if the last arg is a number use it as the page number

	// if the page number was specified then we subtract one from it to make the page index start at 1, then
	// cut the last argument out if it's a number
	if _, err := strconv.Atoi(args[len(args)-1]); err == nil {
		pageNumber = pageNumber - 1
		args = args[:len(args)-1]
	}

	// Joins args so we can search for the whole string
	query := strings.ToLower(strings.Join(args, " "))

	message := ""
	// searches through results for a match
	var listOfMatchingFactoids []string
	for _, o := range factoidsList {
		if strings.Contains(cleanupFactoid(strings.ToLower(o.Content)), query) || strings.Contains(cleanupFactoid(strings.ToLower(o.Name)), query) {
			listOfMatchingFactoids = append(listOfMatchingFactoids, "**"+o.Name+"**```"+cleanupFactoid(o.Content)+"```\n")
		}
	}

	// splits results into groups of the env variable "factoids.max" for the paginator
	var factoidsListPaginated [][]string
	for i := 0; i < len(listOfMatchingFactoids); i += max {
		smallSlice := make([]string, 0, max)
		for j := i; j < i+max && j < len(listOfMatchingFactoids); j++ {
			smallSlice = append(smallSlice, listOfMatchingFactoids[j])
		}

		factoidsListPaginated = append(factoidsListPaginated, smallSlice)
	}

	// if the message is empty let them know nothing was found
	if len(factoidsListPaginated) == 0 {
		err = SendWithSelfDelete(ds, mc.ChannelID, "No matches found.")
		if err != nil {
			return
		}
		return
	}

	// ensures that page number is valid
	if pageNumber < 0 || pageNumber >= len(factoidsListPaginated) {
		err = SendWithSelfDelete(ds, mc.ChannelID, "Page index out of range.")
		if err != nil {
			return
		}
		return
	}

	// combines the specified page (default page 1) into a single string for the description
	message += strings.Join(factoidsListPaginated[pageNumber], "")

	// let the user know total number of pages (provided there are enough results for that)
	footer := ""
	if len(factoidsListPaginated) != 1 {
		footer = "Page " + strconv.Itoa(pageNumber+1) + "/" + strconv.Itoa(len(factoidsListPaginated)) + ". Type !?s " + strings.Join(args, " ") + " " + strconv.Itoa(pageNumber+2) + " to see the next page."
	}

	embed := &discordgo.MessageEmbed{
		Description: message,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
	}

	send := &discordgo.MessageSend{
		Embed: embed,
	}

	_, err = ds.ChannelMessageSendComplex(mc.ChannelID, send)
	if err != nil {
		logger.Err().Printf("Failed to send message\n%s", err)
	}

}

func SendWithSelfDelete(ds *discordgo.Session, channelId, message string) error {
	m, err := ds.ChannelMessageSend(channelId, message)
	if err != nil {
		return err
	}

	go func(ch, id string, session *discordgo.Session) {
		<-time.After(10 * time.Second)
		_ = ds.ChannelMessageDelete(channelId, m.ID)
	}(channelId, m.ID, ds)
	return nil
}

func cleanupFactoid(msg string) string {
	msg = strings.Replace(msg, "[b]", "**", -1)
	msg = strings.Replace(msg, "[/b]", "**", -1)
	msg = strings.Replace(msg, "[u]", "__", -1)
	msg = strings.Replace(msg, "[/u]", "__", -1)
	msg = strings.Replace(msg, "[i]", "*", -1)
	msg = strings.Replace(msg, "[/i]", "*", -1)
	msg = strings.Replace(msg, ";;", "\n", -1)

	return msg
}

type factoid struct {
	Name    string `gorm:"name"`
	Content string `gorm:"content"`
}
