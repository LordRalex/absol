package modules

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
)

var availableModules = make(map[string]api.Module, 0)
var loadedModules = make(map[string]api.Module, 0)

func Load(ds *discordgo.Session, modules []string) {
	if len(modules) == 1 && modules[0] == "all" {
		loadedModules = availableModules
	} else {
		for _, v := range modules {
			logger.Out().Printf("Loading %s\n", v)
			mod := availableModules[v]
			if mod != nil {
				loadedModules[v] = mod
			} else {
				logger.Err().Printf("Module %s does not exist\n", v)
			}
		}
	}

	for k, v := range loadedModules {
		v.Load(ds)
		logger.Out().Printf("Loaded %s\n", k)
	}
}

func Add(module api.Module) {
	availableModules[module.Name()] = module
}

func GetLoaded() map[string]api.Module {
	return loadedModules
}
