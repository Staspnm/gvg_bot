package models

import "time"

type User struct {
	ID            int
	TelegramID    int64
	GameNickname  string
	NineDigitCode string
	GuildName     string
	GuildRole     string // owner, leader, officer, member
	IsActive      bool
	RegisteredAt  time.Time
	LastActivity  time.Time
}
