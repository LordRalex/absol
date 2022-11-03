package log

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/env"
	"github.com/lordralex/absol/api/logger"
	"io"
	"net/http"
	"sync"
)

var lastAuditIds = make(map[string]string)
var auditLastCheck sync.Mutex
var loggedServers []string
var client = &http.Client{}

type Module struct {
	api.Module
}

func (*Module) Load(session *discordgo.Session) {
	loggedServers = env.GetStringArray("log.guilds", ";")

	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnMessageDelete)
	session.AddHandler(OnMessageDeleteBulk)
	session.AddHandler(OnMessageEdit)
	session.AddHandlerOnce(OnConnect)

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages, discordgo.IntentsGuildBans, discordgo.IntentsGuildMembers, discordgo.IntentsMessageContent)
}

func OnConnect(ds *discordgo.Session, mc *discordgo.Connect) {
	auditLastCheck.Lock()
	defer auditLastCheck.Unlock()

	for _, guild := range ds.State.Guilds {
		auditLog, err := ds.GuildAuditLog(guild.ID, "", "", int(discordgo.AuditLogActionMessageDelete), 1)
		if err != nil {
			logger.Err().Printf("Failed to check audit log: %s", err.Error())
		} else {
			for _, v := range auditLog.AuditLogEntries {
				lastAuditIds[guild.ID] = v.ID
			}
		}
	}
}

func downloadAttachment(db *sql.DB, id, url, filename string) {
	if !env.GetBool("log.download") {
		return
	}

	//check to see if URL already exists, if so, skip
	stmt, _ := db.Prepare("SELECT id from attachments WHERE url = ?")
	rows, err := stmt.Query(url)
	_ = stmt.Close()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	hasRows := rows.Next()
	_ = rows.Close()
	if hasRows {
		return
	}

	var data []byte
	response, err := client.Get(url)
	if err == nil {
		defer response.Body.Close()
		data, _ = io.ReadAll(response.Body)
	}

	stmt, _ = db.Prepare("INSERT INTO attachments (message_id, url, name, contents, is_compressed) VALUES (?, ?, ?, COMPRESS(?), 1);")
	err = database.Execute(stmt, id, url, filename, data)
	if err != nil {
		logger.Err().Print(err.Error())
	}
}

func (*Module) Name() string {
	return "log"
}
