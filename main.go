package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/handlers"
	"github.com/lordralex/absol/handlers/servers"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var Session, _ = discordgo.New()

func init() {
	viper.AutomaticEnv()
}

func main() {
	token := viper.GetString("discord_token")

	if token == "" {
		logger.Err().Print("DISCORD_TOKEN must be set in the environment to run this process")
		return
	} else {
		fmt.Printf("Using token: %s\n", token)
	}

	OpenConnection(token)

	//mcf.Schedule(Session)
	servers.Schedule(Session)

	// Wait for a CTRL-C
	fmt.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Clean up
	_ = Session.Close()

	defer func() {
		err := logger.Close()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error closing logger: %s", err.Error())
		}
	}()
}

func OpenConnection(token string) {
	if !strings.HasPrefix(token, "Bot ") {
		token = "Bot " + token
	}
	Session.Token = token

	handlers.RegisterCore(Session)
	handlers.RegisterCommands(Session)

	err := Session.Open()
	if err != nil {
		logger.Err().Print(err.Error())
		os.Exit(1)
	}
}
