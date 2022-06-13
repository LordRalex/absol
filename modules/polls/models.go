package polls

import (
	"gorm.io/gorm"
	"time"
)

type Vote struct {
	gorm.Model
	MessageId string `gorm:"uniqueIndex:vote_idx;index:"`
	UserId    string `gorm:"uniqueIndex:vote_idx"`
	Vote      string
}

type Poll struct {
	gorm.Model
	ChannelId string `gorm:"index"`
	MessageId string `gorm:"uniqueIndex"`
	Title     string
	Closed    bool
	Started   time.Time
	EndAt     time.Time `gorm:"index"`
}
