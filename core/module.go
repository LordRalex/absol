package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/lordralex/absol/modules/alert"
	"github.com/lordralex/absol/modules/cleaner"
	"github.com/lordralex/absol/modules/factoids"
	"github.com/lordralex/absol/modules/hjt"
	"github.com/lordralex/absol/modules/log"
	"github.com/lordralex/absol/modules/mcping"
	"github.com/lordralex/absol/modules/search"
	"github.com/lordralex/absol/modules/twitch"
	"strings"
)

var loadedModules = make(map[string]api.Module, 0)

func LoadModule(ds *discordgo.Session, modules []string) {
	for _, v := range modules {
		logger.Out().Printf("Loading %s\n", v)
		switch strings.ToLower(v) {
		case "twitch":
			loadedModules["twitch"] = &twitch.Module{}
		case "cleaner":
			loadedModules["cleaner"] = &cleaner.Module{}
		case "alert":
			loadedModules["alert"] = &alert.Module{}
		case "log":
			loadedModules["log"] = &log.Module{}
		case "factoids":
			loadedModules["factoids"] = &factoids.Module{}
		case "hjt":
			loadedModules["hjt"] = &hjt.Module{}
		case "search":
			loadedModules["search"] = &search.Module{}
		case "mcping":
			loadedModules["mcping"] = &mcping.Module{}
		default:
			logger.Err().Printf("No module with name %s\n", v)
		}
	}

	for k, v := range loadedModules {
		v.Load(ds)
		logger.Out().Printf("Loaded %s\n", k)
	}
}
