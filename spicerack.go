package spicerack

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"gofig"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
)

type FightWinner int8

const (
	WINNER_RED  FightWinner = 1
	WINNER_BLUE FightWinner = 2
	P1_KEY      string      = "player1name"
	P2_KEY      string      = "player2name"
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

func GetSecretData(sshhh string) (data map[string]string, err error) {
	r, err := http.Get(sshhh)
	if err != nil {
		return
	}

	defer r.Body.Close()
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, &data)
	if err != nil {
		return
	}

	addMrsDash(data)

	return
}

func GofigFromEnv(env string) (*gofig.Config, error) {
	conf_file := os.Getenv(env)
	if len(conf_file) == 0 {
		return nil, errors.New(fmt.Sprintf("%s environment variable has not been set, and needs to point to a JSON configuration file.", env))
	}
	conf, err := gofig.Load(conf_file)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func UpdateFighterElo(red, blue *Fighter, winner FightWinner) {
	red_change := computeScore(red.Elo, blue.Elo)
	blue_change := computeScore(blue.Elo, red.Elo)
	red_k := computeK(red.TotalMatches())
	blue_k := computeK(blue.TotalMatches())

	if winner == WINNER_RED {
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

func addMrsDash(data map[string]string) {
	dash := make([]string, 0)

	if hasFighter(data, "bonegolem") {
		dash = append(dash, "thats_my_boy")
	}

	if len(dash) > 0 {
		data[":mrs_dash"] = strings.Join(dash, "|")
	}
}

func hasFighter(data map[string]string, fighter string) bool {
	lcase_fighter := strings.ToLower(strings.Trim(fighter, " "))
	return strings.ToLower(strings.Trim(data[P1_KEY], " ")) == lcase_fighter ||
		strings.ToLower(strings.Trim(data[P2_KEY], " ")) == lcase_fighter
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
