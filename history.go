package spicerack

import "encoding/json"

type History struct {
	Fighter *Fighter
	Wins    []*FightResult
	Losses  []*FightResult
}

type FightResult struct {
	Opponent   string
	Elo        int
	Victorious bool
}

func (h *History) String() (s string) {
	b, err := json.Marshal(h)
	if err != nil {
		s = ""
	} else {
		s = string(b)
	}
	return
}
