package pastes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
)

type Module struct {
	api.Module
}

func (*Module) Load(_ *discordgo.Session) {
	api.RegisterIntentNeed(discordgo.IntentsGuildMessages)
	if viper.GetString("paste.token") == "" {
		logger.Err().Fatal("Paste token required to use pastes module!")
	}
	if viper.GetString("paste.url") == "" {
		logger.Err().Fatal("Pastebin url required to use pastes module!")
	}
}

func HandleMessage(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if len(mc.Attachments) <= 0 {
		return
	}
	if mc.Attachments[0].ContentType != "text/plain; charset=utf-8" {
		return
	}
	body := &Paste{
		Url: mc.Attachments[0].URL,
		Author: Author{
			Id:            mc.Author.ID,
			Username:      mc.Author.Username,
			Discriminator: mc.Author.Discriminator,
		},
	}
	data := new(bytes.Buffer)
	json.NewEncoder(data).Encode(body)
	req, err := http.NewRequest("POST", viper.GetString("paste.url")+"new", data)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+viper.GetString("paste.token"))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
	defer res.Body.Close()
	location := new(Returned)
	res_body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
	err = json.Unmarshal(res_body, &location)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
	msg := &discordgo.MessageSend{
		Content: "Pastebin'd!",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Emoji: discordgo.ComponentEmoji{
							Name: "📜",
						},
						Label: "View log",
						Style: discordgo.LinkButton,
						URL:   viper.GetString("paste.url") + location.Id,
					},
				},
			},
		},
	}
	_, err = ds.ChannelMessageSendComplex(mc.ChannelID, msg)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
}

type Paste struct {
	Url    string `json:"url"`
	Author Author `json:"author"`
}

type Author struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

type Returned struct {
	Id string `json:"id"`
}
