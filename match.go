package spicerack

import "time"

type Match struct {
	Id       int
	MatchId  int
	RedId    int
	BlueId   int
	RedBets  int
	BlueBets int
	BetCount int
	Winner   int
	Created  time.Time
	Updated  time.Time
}

func (m *Match) InDb() bool {
	return m.Id > 0
}
