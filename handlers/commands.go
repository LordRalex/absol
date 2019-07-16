package handlers

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"strings"
)

var CommandPrefix string
var ClientId string

var Client *http.Client

const TwitchUrl string = "https://api.twitch.tv/helix/users?login="

func RegisterCommands(session *discordgo.Session) {
	CommandPrefix = viper.GetString("prefix")

	if CommandPrefix == "" {
		CommandPrefix = "!?"
	}

	ClientId = viper.GetString("twitch")

	Client = &http.Client{}

	logger.Out().Printf("Adding commands")
	session.AddHandler(OnMessageCommand)
}

func OnMessageCommand(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	if !strings.HasPrefix(mc.Message.Content, CommandPrefix) {
		return
	}

	logger.Out().Printf("Command receieved")

	c, err := ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
		if err != nil {
			logger.Err().Printf("unable to fetch Channel for Message, %s", err)
		} else {
			// Attempt to add this channel into our State
			err = ds.State.ChannelAdd(c)
			if err != nil {
				logger.Err().Printf("error updating State with Channel, %s", err)
			}
		}
	}

	//only work with DMs
	if c == nil || c.Type != discordgo.ChannelTypeDM {
		return
	}

	msg := strings.TrimPrefix(mc.Message.Content, CommandPrefix)

	parts := strings.Split(msg, " ")
	cmd := parts[0]
	args := parts[1:]

	switch strings.ToLower(cmd) {
	case "twitchid":
		{
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

			requestUrl := TwitchUrl + username

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
	}
}
