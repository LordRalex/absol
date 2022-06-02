package polls

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/gorm/clause"
	"strings"
)

func runVoteCast(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	db, err := database.Get()
	if err != nil {
		logger.Err().Println(err.Error())
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Vote failed to be cast..."})
		return
	}

	vote := &Vote{
		MessageId: i.Message.ID,
		UserId:    i.Member.User.ID,
		Vote:      strings.TrimPrefix(i.MessageComponentData().CustomID, "vote:"),
	}

	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"vote"}),
	}).Create(&vote).Error

	if err != nil {
		logger.Err().Println(err.Error())
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Vote failed to be cast..."})
	}

	_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Vote casted!"})
}
