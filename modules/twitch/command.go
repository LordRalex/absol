package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	param := args[0]
	requestUrl := "https://api.twitch.tv/helix/users"

	if cmd == "twitchid" {
		requestUrl += "?login=" + param
	} else if cmd == "twitchname" {
		requestUrl += "?id=" + param
	} else {

	}

	data, err := callTwitch(requestUrl)
	if err != nil {
		logger.Err().Printf("unable to call twitch url %s\n%s", requestUrl, err)
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Failed to get twitch info")
		return
	}

	if data.Data == nil {
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "Failed to get twitch info")
	} else if len(data.Data) == 0 {
		_, _ = ds.ChannelMessageSend(mc.ChannelID, "No such user")
	} else {
		user := data.Data[0]
		_, _ = ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Display Name: %s\nLogin: %s\nID: %s", user.DisplayName, user.Login, user.Id))
	}
}

func callTwitch(requestUrl string) (data TwitchApi, err error) {
	req := &http.Request{}
	req.URL, err = url.Parse(requestUrl)
	if err != nil {
		return
	}

	clientId := viper.GetString("twitch_client_id")

	req.Header = http.Header{}
	req.Method = "GET"
	req.Header.Add("Client-ID", clientId)
	req.Header.Add("Authorization", "Bearer "+accessToken)

	response, err := Client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	err = json.NewDecoder(response.Body).Decode(&data)
	return
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
