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
	if len(args) == 0 {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "You must specify a search string!")
		return
	} else if len(strings.Join(args, "")) < 3 {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Your search is too short!")
		return
	}

	if mc.GuildID != "" {
		_ = factoids.SendWithSelfDelete(ds, mc.ChannelID, "This command may only be used in DMs.")
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
	db.Raw("SELECT * FROM factoids WHERE content LIKE ?", "%"+strings.Join(args, " ")+"%").Scan(&factoidsList)

	// splits results into groups of the env variable "factoids.max" for the paginator
	var factoidsListPaginated [][]factoids.Factoid
	for i := 0; i < len(factoidsList); i += max {
		smallSlice := make([]factoids.Factoid, 0, max)
		for j := i; j < i+max && j < len(factoidsList); j++ {
			smallSlice = append(smallSlice, factoidsList[j])
		}

		factoidsListPaginated = append(factoidsListPaginated, smallSlice)
	}

	// if the message is empty let them know nothing was found
	if len(factoidsListPaginated) == 0 {
		err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "No matches found.")
		if err != nil {
			return
		}
		return
	}

	// ensures that page number is valid
	if pageNumber < 0 || pageNumber >= len(factoidsListPaginated) {
		err = factoids.SendWithSelfDelete(ds, mc.ChannelID, "Page index out of range.")
		if err != nil {
			return
		}
		return
	}

	for _, factoid := range factoidsListPaginated[pageNumber] {
		message += "**" + factoid.Name + "**" + "```" + factoids.CleanupFactoid(factoid.Content) + "```\n"
	}

	footer := ""
	if len(factoidsListPaginated) != 1 {
		footer = "Page " + strconv.Itoa(pageNumber+1) + "/" + strconv.Itoa(len(factoidsListPaginated)) + ". "
		if pageNumber+1 < len(factoidsListPaginated) {
			footer += "Type !?s " + strings.Join(args, " ") + " " + strconv.Itoa(pageNumber+2) + " to see the next page."
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
