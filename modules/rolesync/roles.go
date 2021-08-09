package rolesync

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

func (m *Module) Load(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		syncRoles(ds)

		timer := time.NewTicker(10 * time.Minute)

		for {
			select {
			case <-timer.C:
				{
					syncRoles(ds)
				}}
		}
	}(d)
}

func syncRoles(ds *discordgo.Session) {
	guilds := strings.Split(viper.GetString("ROLESYNC_SERVERS"), ",")

	db, err := database.Get()
	if err != nil {
		return
	}

	for _, guildId := range guilds {
		if guildId == "" {
			continue
		}
		roles, err := ds.GuildRoles(guildId)
		if err != nil {
			logger.Err().Printf("Failed to sync roles for %s: %s\n", guildId, err.Error())
		}
		for _, role := range roles {
			err = db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "position", "permissions"})}).Create(&Role{
				Id:          role.ID,
				GuildId:     guildId,
				Name:        role.Name,
				Position:    role.Position,
				Permissions: role.Permissions,
			}).Error
		}
	}
}

type Role struct {
	Id          string
	GuildId     string `gorm:"column:guild_id"`
	Name        string
	Position    int
	Permissions int
}
