package api

import "github.com/bwmarrin/discordgo"

var intents []discordgo.Intent

func RegisterIntentNeed(neededIntents ...discordgo.Intent) {
	for _, i := range neededIntents {
		add := true
		for _, v := range intents {
			if v == i {
				add = false
				break
			}
		}
		if add {
			intents = append(intents, i)
		}
	}
}

func GetIntent() *discordgo.Intent {
	var intent discordgo.Intent

	for _, v := range intents {
		intent = intent | v
	}

	return discordgo.MakeIntent(intent)
}