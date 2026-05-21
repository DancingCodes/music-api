package main

import "time"

type Music struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"type:varchar(255);index;not null" json:"name"`
	Url        string    `gorm:"type:varchar(1024)" json:"url"`
	PicUrl     string    `gorm:"type:varchar(500)" json:"pic_url"`
	Artists    string    `gorm:"type:varchar(255)" json:"artists"`
	DurationMs int       `gorm:"column:duration_ms;type:int" json:"duration_ms"`
	Lyric      string    `gorm:"type:text" json:"lyric"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
