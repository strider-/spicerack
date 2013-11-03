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

func (fc *FightCard) TakingBets() bool {
	return fc.Status == "open"
}

func (fc *FightCard) InProgress() bool {
	return fc.Status == "locked"
}

func (fc *FightCard) WeHaveAWinner() bool {
	return fc.Status == "1" || fc.Status == "2"
}

func (fc *FightCard) Winner() string {
	if fc.Status == "1" {
		return fc.RedName
	} else if fc.Status == "2" {
		return fc.BlueName
	}
	return ""
}

func (fc *FightCard) Upset(factor float64) bool {
	return (fc.Status == "1" && (float64(fc.BlueTotal)*factor) > float64(fc.RedTotal)) ||
		(fc.Status == "2" && (float64(fc.RedTotal)*factor) > float64(fc.BlueTotal))
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
