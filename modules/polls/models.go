package polls

import "gorm.io/gorm"

type Vote struct {
	gorm.Model
	MessageId string `gorm:"uniqueIndex:vote_idx;index:"`
	UserId    string `gorm:"uniqueIndex:vote_idx"`
	Vote      string
}

type Poll struct {
	gorm.Model
	MessageId string `gorm:"index:,unique"`
	Title     string
	Closed    bool
}
