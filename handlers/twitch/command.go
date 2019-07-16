package twitch

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

var ClientId string
var Client *http.Client

const ApiUrl string = "https://api.twitch.tv/helix/users?login="

func init() {
	ClientId = viper.GetString("twitch")

	Client = &http.Client{}
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, c *discordgo.Channel, cmd string, args []string) {
	var err error

	if ClientId == "" {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to get twitch info, contact the admin")
		logger.Err().Printf("Token for twitch is not configured")
		return
	}

	if len(args) != 1 {
		logger.Err().Printf("Name required")
		return
	}

	username := args[0]

	requestUrl := ApiUrl + username

	req := &http.Request{}
	req.URL, err = url.Parse(requestUrl)
	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Username does not seem like it's valid")
		logger.Err().Printf("unable to parse url %s\n%s", requestUrl, err)
		return
	}

	req.Header = http.Header{}
	req.Method = "GET"
	req.Header.Add("Client-ID", ClientId)

	response, err := Client.Do(req)
	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to get twitch info, contact the admin")
		logger.Err().Printf("unable to call twitch API\n%s", err)
		return
	}
	defer response.Body.Close()

	data := &TwitchApi{}
	_ = json.NewDecoder(response.Body).Decode(data)
	if data.Data == nil {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to get twitch info, contact the admin")
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	} else if len(data.Data) == 0 {
		_, err = ds.ChannelMessageSend(c.ID, "No such user called "+username)
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	} else {
		_, err = ds.ChannelMessageSend(c.ID, username+": "+data.Data[0].Id)
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	}
}
