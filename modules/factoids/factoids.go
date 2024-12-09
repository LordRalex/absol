package factoids

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"slices"
	"strings"
	"time"
)

var factoidsUrl string
var maxFactoids int

type Module struct {
	api.Module
}

func (*Module) Load(ds *discordgo.Session) {
	api.RegisterCommand("f", RunCommand)
	api.RegisterCommand("factoid", RunCommand)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsDirectMessages)

	factoidsUrl = env.GetOr("hjt.url", "https://minecrafthopper.net/factoids.json")

	maxFactoids = env.GetInt("factoids.max")
	if maxFactoids <= 0 {
		maxFactoids = 5
	}
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, cmd string, args []string) {
	if len(args) == 0 {
		return
	}

	guilds := env.GetStringArray("factoids.guilds", ";")
	if !slices.Contains(guilds, mc.GuildID) {
		return
	}

	factoids := make([]string, 0)
	if cmd == "" {
		factoids = []string{strings.ToLower(cmd)}
	}

	if len(mc.MentionRoles)+len(mc.MentionChannels) > 0 {
		_ = SendWithSelfDelete(ds, mc.ChannelID, "Cannot mention to roles or channels")
		return
	}

	for _, v := range args {
		skip := false
		for _, m := range mc.Mentions {
			if "<@!"+m.ID+">" == v || "<@"+m.ID+">" == v {
				skip = true
				break
			}
		}
		if !skip {
			factoids = append(factoids, strings.ToLower(v))
		}
	}

	if len(factoids) > maxFactoids {
		_ = SendWithSelfDelete(ds, mc.ChannelID, fmt.Sprintf("Cannot send more than %d factoids at once", maxFactoids))
		return
	}

	matches, err := getMatchingFactoids(factoids)
	if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		return
	} else if len(matches) == 0 {
		_ = SendWithSelfDelete(ds, mc.ChannelID, "No factoid with the given name was found: "+strings.Join(factoids, ", "))
		return
	}

	if len(factoids) != len(matches) {
		//we have a missing one...
		missing := make([]string, 0)
		for _, v := range factoids {
			good := false
			for _, k := range matches {
				if match(k.Name, v) {
					good = true
					break
				}
			}
			if !good {
				missing = append(missing, v)
			}
		}

		//someone is dumb..... and put the same factoid twice
		if len(missing) != 0 {
			_ = SendWithSelfDelete(ds, mc.ChannelID, "No factoid with the given name(s) was found: "+strings.Join(missing, ", "))
			return
		}
	}

	msg := ""
	for i, v := range factoids {
		for _, o := range matches {
			if match(v, o.Name) {
				msg += CleanupFactoid(o.Content)
				if i+1 != len(factoids) {
					msg += "\n\n"
				}
			}
		}
	}

	header := ""
	if len(mc.Message.Mentions) > 0 || mc.MessageReference != nil {
		//the golang set
		mentions := make(map[string]bool, 0)

		//if we have an @, we'll add it to the message
		for _, v := range mc.Mentions {
			mentions[v.ID] = true
		}

		if mc.MessageReference != nil && mc.MessageReference.MessageID != "" {
			replyMsg, err := ds.ChannelMessage(mc.MessageReference.ChannelID, mc.MessageReference.MessageID)
			if err == nil && replyMsg.ID != "" {
				mentions[replyMsg.Author.ID] = true
			}
		}

		for k := range mentions {
			header += "<@" + k + "> " + " "
		}

		header += "Please refer to the below information."
	}

	embed := &discordgo.MessageEmbed{
		Description: msg,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "I am a bot, I will not respond to you. Command issued by " + mc.Author.Username,
		},
	}

	send := &discordgo.MessageSend{
		Content: header,
		Embeds:  []*discordgo.MessageEmbed{embed},
	}

	if env.GetBool("factoids.delete") {
		_ = ds.ChannelMessageDelete(mc.ChannelID, mc.ID)
	}

	_, err = ds.ChannelMessageSendComplex(mc.ChannelID, send)
	if err != nil {
		logger.Err().Printf("Failed to send message\n%s", err)
	}
}

func SendWithSelfDelete(ds *discordgo.Session, channelId, message string) error {
	m, err := ds.ChannelMessageSend(channelId, message)
	if err != nil {
		return err
	}

	go func(ch, id string, session *discordgo.Session) {
		<-time.After(10 * time.Second)
		_ = ds.ChannelMessageDelete(channelId, m.ID)
	}(channelId, m.ID, ds)
	return nil
}

func CleanupFactoid(msg string) string {
	msg = strings.Replace(msg, "[b]", "**", -1)
	msg = strings.Replace(msg, "[/b]", "**", -1)
	msg = strings.Replace(msg, "[u]", "__", -1)
	msg = strings.Replace(msg, "[/u]", "__", -1)
	msg = strings.Replace(msg, "[i]", "*", -1)
	msg = strings.Replace(msg, "[/i]", "*", -1)
	msg = strings.Replace(msg, ";;", "\n", -1)

	if strings.Contains(msg, "https://") || strings.Contains(msg, "http://") {
		msgsplit := strings.Split(msg, " ")
		for k, v := range msgsplit {
			if strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "http://") {
				msgsplit[k] = "<" + v + ">"
			}
		}
		msg = strings.Join(msgsplit, " ")
	}

	return msg
}

type Factoid struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (*Module) Name() string {
	return "factoids"
}

func getMatchingFactoids(names []string) ([]Factoid, error) {
	data, err := api.GetFromUrl(factoidsUrl)
	if err != nil {
		return nil, err
	}
	var factoids []Factoid
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&factoids)
	if err != nil {
		return nil, err
	}

	var matches []Factoid

	for _, v := range names {
		for _, z := range factoids {
			if match(v, z.Name) {
				matches = append(matches, z)
			}
		}
	}

	return matches, err
}

func match(v, z string) bool {
	return strings.ToLower(v) == strings.ToLower(z)
}
