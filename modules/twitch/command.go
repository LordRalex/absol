package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/twitch"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var Client *http.Client

const ApiUrl string = "https://api.twitch.tv/helix/users?login="

var locker sync.RWMutex
var accessToken string
var timeout time.Time

func init() {
	Client = &http.Client{}
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, cmd string, args []string) {
	var err error

	if len(args) != 1 {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Name required")
		logger.Err().Printf("Name required")
		return
	}

	err = refreshToken()

	locker.RLock()
	defer locker.RUnlock()

	username := args[0]

	requestUrl := ApiUrl + username

	req := &http.Request{}
	req.URL, err = url.Parse(requestUrl)
	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Username does not seem like it's valid")
		logger.Err().Printf("unable to parse url %s\n%s", requestUrl, err)
		return
	}

	clientId := viper.GetString("twitch_client_id")

	req.Header = http.Header{}
	req.Method = "GET"
	req.Header.Add("Client-ID", clientId)
	req.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := Client.Do(req)
	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to get twitch info, contact the admin")
		logger.Err().Printf("unable to call twitch API\n%s", err)
		return
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	data := &TwitchApi{}
	err = json.NewDecoder(response.Body).Decode(data)

	if err != nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to get twitch info, contact the admin")
		logger.Err().Printf("unable to call twitch API\n%s", err)
		return
	}

	if data.Data == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to get twitch info, contact the admin")
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	} else if len(data.Data) == 0 {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "No such user called "+username)
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	} else {
		_, err = ds.ChannelMessageSend(mc.ChannelID, username+": "+data.Data[0].Id)
		if err != nil {
			logger.Err().Printf("unable to call twitch API\n%s", err)
			return
		}
	}
}

func refreshToken() error {
	if timeout.Before(time.Now()) {
		locker.Lock()
		defer locker.Unlock()

		//if we had 2 requests hit at the same time, validate it's okay now
		if timeout.After(time.Now()) {
			return nil
		}

		clientId := viper.GetString("twitch_client_id")
		clientSecret := viper.GetString("twitch_client_secret")

		if clientId == "" || clientSecret == "" {
			logger.Err().Printf("Token for twitch is not configured")
			return errors.New("token not configured")
		}

		oauth2Config := &clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     twitch.Endpoint.TokenURL,
			Scopes:       []string{"user_read"},
		}

		token, err := oauth2Config.Token(context.Background())
		if err == nil {
			accessToken = token.AccessToken
			//drop timeout to be 1 minute before this expires
			timeout = token.Expiry.Add(time.Minute * -1)
		}

		return err
	}

	return nil
}
