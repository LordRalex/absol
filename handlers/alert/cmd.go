package alert

import (
	"encoding/json"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"github.com/satori/go.uuid"
	"strings"
)

func ImportFromDatabase() {
	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return
	}

	var data []string
	err = db.Table("sites_timed_out").Pluck("log", &data).Error
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return
	}

	for _, d := range data {
		r := strings.NewReader(d)

		var m map[string]interface{}

		err = json.NewDecoder(r).Decode(&m)
		if err != nil {
			logger.Err().Printf("Error decoding: %s\n", err.Error())
			continue
		}

		id := uuid.NewV4().String()
		err = submitToElastic(id, m)
		if err != nil {
			logger.Err().Printf("Error sending to ES: %s\n", err.Error())
		}
	}
}
