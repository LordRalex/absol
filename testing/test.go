package main

import (
	"github.com/lordralex/absol/handlers/mcf"
	"github.com/spf13/viper"
)

func main () {
	viper.AutomaticEnv()
	mcf.RunTick()
}
