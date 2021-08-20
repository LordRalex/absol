package search

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/lordralex/absol/modules/factoids"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

type Module struct {
	api.Module
}

// Load absol commands API
func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("search", RunCommand)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsDirectMessages)
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, _ string, args []string) {
	if mc.GuildID != "" {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "This command may only be used in DMs.")
		return
	}

	if len(args) == 0 {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "You must specify a search string!")
		return
	} else if len(strings.Join(args, "")) < 3 {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Your search is too short!")
		return
	}

	max := viper.GetInt("factoids.max")
	if max == 0 {
		max = 5
	}

	db, err := database.Get()
	if err != nil {
		err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
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

	message := ""
	// searches through results for a match
	// gets the factoids table
	var factoidsList []factoids.Factoid
	var rows int64
	db.Where("content LIKE ? OR name LIKE ?", "%"+strings.Join(args, " ")+"%", "%"+strings.Join(args, " ")+"%").Offset(pageNumber*max + 1).Limit(max).Find(&factoidsList).Count(&rows)

	// if the message is empty let them know nothing was found
	if len(factoidsList) == 0 {
		err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "No matches found.")
		if err != nil {
			return
		}
		return
	}

	// ensures that page number is valid
	if pageNumber < 0 || pageNumber >= len(factoidsList) {
		err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Page index out of range.")
		if err != nil {
			return
		}
		return
	}

	for _, factoid := range factoidsList {
		message += "**" + factoid.Name + "**" + "```" + factoids.CleanupFactoid(factoid.Content) + "```\n"
	}

	footer := ""
	if len(factoidsList) != 1 {
		footer = "Page " + strconv.Itoa(pageNumber+1) + "/" + strconv.Itoa(int(rows)/max) + ". "
		if pageNumber+1 < len(factoidsList) {
			footer += "Type !?search " + strings.Join(args, " ") + " " + strconv.Itoa(pageNumber+2) + " to see the next page."
		}
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
