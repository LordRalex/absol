package hjt

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
)

type Module struct {
	api.Module
}

var hjtUrl string = env.GetOr("hjt.url", "https://minecrafthopper.net/hjt.json")

func (*Module) Load(ds *discordgo.Session) {
	guilds := api.CleanStringArray(env.GetStringArray("hjt.guilds", ";"))

	api.RegisterSlashCommnand(ds, guilds, hjtOperation, runCommand)
}

var hjtOperation = &discordgo.ApplicationCommand{
	Name:        "hjt",
	Description: "Checks a HJT report against a list of known problematic programs",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{

		{
			Name:        "url",
			Description: "URL to paste",
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

	pasteLink := i.ApplicationCommandData().Options[0].StringValue()

	content, err := api.GetFromUrl(pasteLink)
	if err != nil {
		msg := "Invalid URL"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	var values []HJT
	values, err = getHjts()
	if err != nil {
		logger.Err().Printf("Failed to connect to database\n%s", err)
		msg := "Failed to connect to database"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	var results []HJT
	var matches bool
	for _, v := range values {
		matches, err = regexp.Match(v.MatchCriteria, content)
		if err != nil {
			msg := v.Name + " is an invalid regex statement: " + err.Error()
			_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			return
		}
		if matches {
			results = append(results, v)
		}
	}

	if len(results) == 0 {
		msg := "Report for " + pasteLink + "\nNo matches found"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	message := "Report for " + pasteLink + "\n"
	for id, v := range results {
		if id != 0 {
			message += "\n"
		}
		message += v.GetEmojiString() + " [" + v.Category + "] " + v.Name + ": " + v.Description
	}
	_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

type HJT struct {
	Id            uint
	Name          string
	MatchCriteria string `json:"match_criteria"`
	Description   string
	Category      string
	Severity      Severity
}

type Severity int

var SeverityInfo Severity = 0
var SeverityLow Severity = 1
var SeverityMedium Severity = 2
var SeverityHigh Severity = 3

func (s HJT) GetEmojiString() string {
	switch s.Severity {
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

func (*Module) Name() string {
	return "hjt"
}

func getHjts() ([]HJT, error) {
	data, err := api.GetFromUrl(hjtUrl)
	if err != nil {
		return nil, err
	}
	var results []HJT
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&results)
	return results, err
}
