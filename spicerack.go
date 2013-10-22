package spicerack

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math"
)

type FightWinner int8

const (
	WinnerRed  FightWinner = 1
	WinnerBlue FightWinner = 2
)

func Db(user, password, dbname string) *Repository {
	return &Repository{
		user:        user,
		password:    password,
		dbname:      dbname,
		single_conn: false,
		db:          nil,
	}
}

func OpenDb(user, password, dbname string) (*Repository, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname))
	if err != nil {
		return nil, err
	}
	return &Repository{
		user:        user,
		password:    password,
		dbname:      dbname,
		single_conn: true,
		db:          db,
	}, nil
}

func UpdateFighterElo(red, blue *Fighter, winner FightWinner) {
	red_change := computeScore(red.Elo, blue.Elo)
	blue_change := computeScore(blue.Elo, red.Elo)
	red_k := computeK(red.TotalMatches())
	blue_k := computeK(blue.TotalMatches())

	if winner == WinnerRed {
		red.Win += 1
		blue.Loss += 1
		red.Elo += int(red_k * (1 - red_change))
		blue.Elo += int(blue_k * (0 - blue_change))
	} else {
		red.Loss += 1
		blue.Win += 1
		red.Elo += int(red_k * (0 - red_change))
		blue.Elo += int(blue_k * (1 - blue_change))
	}
}

func computeScore(self, opponent int) float64 {
	return 1.0 / (1.0 + (math.Pow10((opponent - self) / 400)))
}

func computeK(match_count int) (k float64) {
	k = 0
	if match_count <= 10 {
		k = 100
	}
	if match_count <= 30 && match_count > 10 {
		k = 50
	}
	if match_count > 30 {
		k = 25
	}
	return
}
