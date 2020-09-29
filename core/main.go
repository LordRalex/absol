package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var Session *discordgo.Session

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	modules := os.Args[1:]

	token := viper.GetString("discord_token")

	if token == "" {
		logger.Err().Print("DISCORD_TOKEN must be set in the environment to run this process")
		return
	} else {
		fmt.Printf("Using token: %s\n", token)
	}

	defer func() {
		err := logger.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error closing logger: %s", err.Error())
		}
	}()

	defer database.Close()

	Session, _ = discordgo.New()
	defer Session.Close()

	if len(modules) > 0 {
		LoadModule(Session, modules)
	}

	OpenConnection(token)

	// Wait for a CTRL-C
	fmt.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Shutting down")
}

func OpenConnection(token string) {
	if !strings.HasPrefix(token, "Bot ") {
		token = "Bot " + token
	}
	Session.Token = token
	Session.Identify.Intents = api.GetIntent()

	EnableCommands(Session)

	err := Session.Open()
	if err != nil {
		logger.Err().Print(err.Error())
		os.Exit(1)
	}
}
