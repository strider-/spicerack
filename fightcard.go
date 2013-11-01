package spicerack

import (
	"fmt"
	"strings"
)

type FightCard struct {
	RedName   string `json:"player1name"`
	BlueName  string `json:"player2name"`
	RedTotal  int    `json:"player1total,string"`
	BlueTotal int    `json:"player2total,string"`
	Status    string `json:"status"`
	MrsDash   string
}

func (fc *FightCard) Upset() bool {
	if fc.Status == "1" && fc.BlueTotal > fc.RedTotal {
		return true
	} else if fc.Status == "2" && fc.RedTotal > fc.BlueTotal {
		return true
	}
	return false
}

func (fc *FightCard) Odds() string {
	if fc.Status != "open" {
		if fc.RedTotal > fc.BlueTotal {
			return fmt.Sprintf("%2.1f:1", float64(fc.RedTotal)/float64(fc.BlueTotal))
		} else {
			return fmt.Sprintf("1:%2.1f", float64(fc.BlueTotal)/float64(fc.RedTotal))
		}
	}
	return ""
}

func (fc *FightCard) Involves(fighter string) bool {
	return strings.ToLower(fc.RedName) == strings.ToLower(fighter) ||
		strings.ToLower(fc.BlueName) == strings.ToLower(fighter)
}
