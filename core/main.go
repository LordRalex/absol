package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/lordralex/absol/modules"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var Session *discordgo.Session

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	m := os.Args[1:]

	token := viper.GetString("discord.token")

	if token == "" {
		logger.Err().Print("DISCORD_TOKEN must be set in the environment to run this process")
		return
	}

	defer func() {
		err := logger.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error closing logger: %s", err.Error())
		}
	}()

	if !strings.HasPrefix(token, "Bot ") {
		token = "Bot " + token
		viper.Set("discord.token", token)
	}

	SetApplicationId()

	Session, _ = discordgo.New(token)
	defer Session.Close()

	modules.Load(Session, m)

	OpenConnection()

	// Wait for a CTRL-C
	fmt.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Shutting down")
}

func OpenConnection() {
	Session.Identify.Intents = api.GetIntent()

	EnableCommands(Session)

	err := Session.Open()
	if err != nil {
		logger.Err().Print(err.Error())
		os.Exit(1)
	}
}

func SetApplicationId() {
	u, err := url.Parse("https://discord.com/api/oauth2/applications/@me")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	request := &http.Request{
		Method: "GET",
		URL:    u,
		Header: map[string][]string{
			"Authorization": {viper.GetString("discord.token")},
		},
	}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var app discordgo.Application
	json.NewDecoder(response.Body).Decode(&app)
	viper.Set("app.id", app.ID)
}
