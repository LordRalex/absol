package polls

type Vote struct {
	MessageId string `gorm:"uniqueIndex:vote_idx;index:"`
	UserId    string `gorm:"uniqueIndex:vote_idx"`
	Vote      string
}

type Poll struct {
	MessageId string `gorm:"index:,unique"`
	Title     string
	Closed    bool
}
