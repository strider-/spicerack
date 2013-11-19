package spicerack

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JKallhoff/gofig"
	_ "github.com/lib/pq"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

type FightWinner int8

const (
	WINNER_RED  FightWinner = 1
	WINNER_BLUE FightWinner = 2

	TIER_R int = 0 // Retired
	TIER_S int = 1
	TIER_A int = 2
	TIER_B int = 3
	TIER_P int = 4 // Potato
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

func LogIntoSaltyBet(user, pass string) (client *http.Client, err error) {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}
	r, err := client.PostForm("http://www.saltybet.com/authenticate?signin=1", url.Values{
		"email":        {user},
		"pword":        {pass},
		"authenticate": {"signin"},
	})
	if err == nil {
		defer r.Body.Close()
		page, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(page), "ui-state-error") {
			err = errors.New("Invalid saltybet credentials")
		}
	}
	return
}

func GetSecretData(sshhh string) (fc *FightCard, err error) {
	raw, err := getJson(sshhh)
	if err != nil {
		return
	}

	fc = &FightCard{}
	err = json.Unmarshal(raw, fc)
	if err != nil {
		return
	}

	addMrsDash(fc)

	return
}

func GetFighterStats(client *http.Client, url string) (fs *FighterStats, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	raw, _ := ioutil.ReadAll(resp.Body)

	fs = &FighterStats{}
	err = json.Unmarshal(raw, fs)
	return
}

func getJson(url string) (raw []byte, err error) {
	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()

	raw, err = ioutil.ReadAll(r.Body)
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

func addMrsDash(fc *FightCard) {
	dash := make([]string, 0)

	if fc.Involves("bonegolem") {
		dash = append(dash, "thats_my_boy")
	}
	if fc.Involves("sissy") {
		dash = append(dash, "fake_astro")
	}
	if fc.Involves("daimon 71113") {
		dash = append(dash, "the_gawd")
	}

	fc.MrsDash = dash
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
