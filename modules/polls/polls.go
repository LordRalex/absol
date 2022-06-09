package polls

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

type Module struct {
	api.Module
}

var appId string
var client = &http.Client{}

func (*Module) Load(ds *discordgo.Session) {
	appId = viper.GetString("app.id")

	var guilds []string

	maps := strings.Split(viper.GetString("POLLS_GUILDS"), ";")
	for _, v := range maps {
		if v == "" {
			continue
		}

		guilds = append(guilds, v)
	}

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, guild := range guilds {
			for _, v := range []*discordgo.ApplicationCommand{createPollOperation, updatePollOperation, closePollOperation} {
				logger.Out().Printf("Registering %s for guild %s\n", v.Name, guild)
				_, err := s.ApplicationCommandCreate(appId, guild, v)
				if err != nil {
					logger.Err().Printf("Cannot create slash command %q: %v", v.Name, err)
				}
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
				if i.ApplicationCommandData().Name == createPollOperation.Name {
					runCreateCommand(s, i)
				}
				if i.ApplicationCommandData().Name == updatePollOperation.Name {
					runUpdateCommand(s, i)
				}
				if i.ApplicationCommandData().Name == closePollOperation.Name {
					runCloseCommand(s, i)
				}
			}
		case discordgo.InteractionMessageComponent:
			{
				if strings.HasPrefix(i.Interaction.MessageComponentData().CustomID, "vote:") {
					_ = ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Flags: uint64(discordgo.MessageFlagsEphemeral)},
					})

					runVoteCast(ds, i)
				}
			}
		}
	})

	db, err := database.Get()
	if err != nil {
		logger.Err().Println(err.Error())
	}

	err = db.AutoMigrate(&Poll{}, &Vote{})
	if err != nil {
		logger.Err().Println(err.Error())
	}
}

func downloadFile(url string) (data string, err error) {
	response, err := client.Get(url)

	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if err != nil {
		return "", err
	}

	limited := io.LimitReader(response.Body, 1024*1024) //1MB
	d, err := io.ReadAll(limited)
	return string(d), err
}

func splitToRows(choices []string) []discordgo.MessageComponent {
	limit := 5

	components := make([]discordgo.MessageComponent, 0)
	row := discordgo.ActionsRow{}

	for _, v := range choices {
		row.Components = append(row.Components, discordgo.Button{CustomID: "vote:" + v, Style: discordgo.PrimaryButton, Label: v})

		if len(row.Components) == limit {
			components = append(components, row)
			row = discordgo.ActionsRow{}
		}
	}

	if len(row.Components) > 0 {
		components = append(components, row)
	}

	return components
}

func hasDupes(choices []string) bool {
	for k, v := range choices {
		index := k + 1

		for ; index < len(choices); index++ {
			if v == choices[index] {
				return true
			}
		}
	}

	return false
}

func (Module) Name() string {
	return "polls"
}
