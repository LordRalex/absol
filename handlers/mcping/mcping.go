package mcping

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"os/exec"
)

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, c *discordgo.Channel, cmd string, args []string) {
	if len(args) == 0 {
		return
	}

	server := args[0]

	pythonCmd := exec.Command("python", "mcping.py", server)
	out, err := pythonCmd.Output()

	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to run script")
		logger.Err().Printf("Failed to run script\n%s", err)
		return
	}

	_, err = ds.ChannelMessageSend(c.ID, string(out))
}
