package models

import "time"

type BattleResult struct {
	ID         int
	UserID     int
	Location   string // T1, T2, ..., F2
	EnemySquad string
	OwnSquad   string
	FlagsCount int
	ReportedAt time.Time
	GuildName  string
	BattleDate time.Time
}
