package hjt

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/gorm"
	"io"
	"net/http"
	"regexp"
)

type Module struct {
	api.Module
}

var appId string

func (*Module) Load(ds *discordgo.Session) {
	appId = env.Get("discord.app_id")

	var guilds []string

	maps := env.GetStringArray("hjt.guilds", ";")
	for _, v := range maps {
		if v == "" {
			continue
		}

		guilds = append(guilds, v)
	}

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, v := range guilds {
			logger.Out().Printf("Registering %s for guild %s\n", hjtOperation.Name, v)
			_, err := s.ApplicationCommandCreate(appId, v, hjtOperation)
			if err != nil {
				logger.Err().Printf("Cannot create slash command %q: %v", hjtOperation.Name, err)
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
				if i.ApplicationCommandData().Name == hjtOperation.Name {
					runCommand(s, i)
				}
			}
		}
	})
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

	content, err := readFromUrl(pasteLink)
	if err != nil {
		msg := "Invalid URL"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Failed to connect to database\n%s", err)
		msg := "Failed to connect to database"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	var values []HJT
	err = db.Table("hjts").Select([]string{"id", "name", "match_criteria"}).Find(&values).Error

	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		msg := "Failed to connect to database"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	var results []uint
	for _, v := range values {
		matches, err := regexp.Match(v.MatchCriteria, content)
		if err != nil {
			msg := v.Name + " is an invalid regex statement: " + err.Error()
			_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			return
		}
		if matches {
			results = append(results, v.Id)
		}
	}

	if len(results) == 0 {
		msg := "Report for " + pasteLink + "\nNo matches found"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	var data []HJT
	err = db.Find(&data, results).Order("severity desc").Error
	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		msg := "Failed to connect to database"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	message := "Report for " + pasteLink + "\n"
	for i, v := range data {
		if i != 0 {
			message += "\n"
		}
		message += v.SeverityEmoji + " [" + v.Category + "] " + v.Name + ": " + v.Description
	}
	_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

func readFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

type HJT struct {
	Id            uint
	Name          string
	MatchCriteria string
	Description   string
	Category      string
	Severity      Severity
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

func (*Module) Name() string {
	return "hjt"
}
