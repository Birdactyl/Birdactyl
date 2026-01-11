package models

type Setting struct {
	Key   string `gorm:"type:varchar(255);primaryKey" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
