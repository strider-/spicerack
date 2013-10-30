package spicerack

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type RematchState int8

const (
	Unknown     RematchState = -1
	NeverFought RematchState = 0
	RedBeatBlue RematchState = 1
	BlueBeatRed RematchState = 2
	TradedWins  RematchState = 3
)

type Repository struct {
	user, password, dbname string
	single_conn            bool
	db                     *sql.DB
}

func (r *Repository) GetFighters(red, blue string) (lF, rF *Fighter, err error) {
	criteria := "Name=$1"
	return r.getFighters(red, blue, criteria)
}

func (r *Repository) SearchFighters(red, blue string) (lF, rF *Fighter, err error) {
	criteria := "lower(Name)=lower($1) AND tier > 0"
	return r.getFighters(red, blue, criteria)
}

func (r *Repository) SearchRetiredFighters(red, blue string) (lF, rF *Fighter, err error) {
	criteria := "lower(Name)=lower($1) AND tier = 0"
	return r.getFighters(red, blue, criteria)
}

func (r *Repository) getFighters(red, blue, criteria string) (lF, rF *Fighter, err error) {
	db, err := r.open()
	if err != nil {
		return
	}
	defer r.close(db)

	sql := fmt.Sprintf("SELECT Id, Name, Wins, Losses, Elo, Total_Bets, Character_Id, Tier, Created_At, Updated_At FROM Champions WHERE %s", criteria)

	lF = &Fighter{}
	err = db.QueryRow(sql, red).Scan(&lF.Id, &lF.Name, &lF.Win, &lF.Loss, &lF.Elo, &lF.TotalBets, &lF.CharacterId, &lF.Tier, &lF.Created, &lF.Updated)
	if err != nil {
		lF = nil
		err = nil
	}

	rF = &Fighter{}
	err = db.QueryRow(sql, blue).Scan(&rF.Id, &rF.Name, &rF.Win, &rF.Loss, &rF.Elo, &rF.TotalBets, &rF.CharacterId, &rF.Tier, &rF.Created, &rF.Updated)
	if err != nil {
		rF = nil
		err = nil
	}

	return
}

func (r *Repository) GetRematchState(red, blue *Fighter) (RematchState, error) {
	if red == nil || blue == nil {
		return NeverFought, nil
	}

	db, err := r.open()
	if err != nil {
		return Unknown, err
	}
	defer r.close(db)

	sql := `SELECT 
                red_champion_id AS Red, blue_champion_id AS Blue, Winner 
            FROM Fights 
            WHERE 
                (red_champion_id=$1 AND blue_champion_id=$2) OR (red_champion_id=$2 AND blue_champion_id=$1)`

	rows, err := db.Query(sql, red.Id, blue.Id)
	if err != nil {
		return Unknown, err
	}

	wins := map[int]int{
		red.Id:  0,
		blue.Id: 0,
	}

	for rows.Next() {
		var r, b, winner int
		rows.Scan(&r, &b, &winner)

		if winner == 1 {
			wins[r] += 1
		} else if winner == 2 {
			wins[b] += 1
		}
	}

	if wins[red.Id] > 0 && wins[blue.Id] > 0 {
		return TradedWins, nil
	} else if wins[blue.Id] > 0 {
		return BlueBeatRed, nil
	} else if wins[red.Id] > 0 {
		return RedBeatBlue, nil
	}

	return NeverFought, nil
}

func (r *Repository) MatchExists(id int) bool {
	db, _ := r.open()
	defer r.close(db)
	result, _ := db.Exec("SELECT id FROM Fights WHERE match_id=$1", id)
	rows, _ := result.RowsAffected()
	return rows > 0
}

func (r *Repository) InsertMatch(m *Match) error {
	db, err := r.open()
	if err != nil {
		return err
	}
	defer r.close(db)
	err = db.QueryRow(`
        INSERT INTO Fights 
            (red_champion_id, blue_champion_id, bets_red, bets_blue, bet_count, winner, created_at, updated_at, match_id) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		m.RedId, m.BlueId, m.RedBets, m.BlueBets, m.BetCount, m.Winner, m.Created, m.Updated, m.MatchId).Scan(&m.Id)
	return err
}

func (r *Repository) GetFighter(nameOrCharId interface{}) (f *Fighter, err error) {
	db, _ := r.open()
	defer r.close(db)
	f = &Fighter{
		Id:  0,
		Elo: 300, Win: 0, Loss: 0, TotalBets: 0,
		CharacterId: 0, Tier: 0,
		Created: time.Now(), Updated: time.Now(),
	}
	baseSql := "SELECT id, name, elo, wins, losses, total_bets, character_id, tier, created_at, updated_at FROM Champions WHERE %s"

	var sql string = ""
	switch nameOrCharId.(type) {
	case string:
		f.Name = nameOrCharId.(string)
		sql = fmt.Sprintf(baseSql, "name=$1")
	case int, int8, int16, int64:
		sql = fmt.Sprintf(baseSql, "character_id=$1")
	default:
		f = nil
		err = errors.New("GetFighter needs a string (name) or int (character_id) to fetch by.")
		return
	}

	row := db.QueryRow(sql, nameOrCharId)
	row.Scan(&f.Id, &f.Name, &f.Elo, &f.Win, &f.Loss, &f.TotalBets, &f.CharacterId, &f.Tier, &f.Created, &f.Updated)
	return
}

func (r *Repository) UpdateFighterInTrans(f *Fighter, tx *sql.Tx) error {
	return r.updateFighter(f, tx)
}

func (r *Repository) UpdateFighter(f *Fighter) error {
	if r.single_conn {
		return r.updateFighter(f, r.db)
	} else {
		db, _ := r.open()
		defer r.close(db)
		return r.updateFighter(f, db)
	}
}

func (r *Repository) updateFighter(f *Fighter, x interface {
	QueryRow(string, ...interface{}) *sql.Row
	Exec(string, ...interface{}) (sql.Result, error)
}) error {
	if !f.InDb() {
		err := x.QueryRow("INSERT INTO Champions (name, elo, wins, losses, total_bets, character_id, tier, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
			f.Name, f.Elo, f.Win, f.Loss, f.TotalBets, f.CharacterId, f.Tier, f.Created, f.Updated).Scan(&f.Id)
		if err != nil {
			return err
		}
	} else {
		f.Updated = time.Now()
		_, err := x.Exec("UPDATE Champions SET elo=$1, wins=$2, losses=$3, total_bets=$4, character_id=$5, tier=$6, updated_at=$7 WHERE id=$8",
			f.Elo, f.Win, f.Loss, f.TotalBets, f.CharacterId, f.Tier, f.Updated, f.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ResetElo(base int) error {
	db, _ := r.open()
	defer db.Close()

	_, err := db.Exec("UPDATE Champions SET Elo=$1, wins=0, losses=0, total_bets=0", base)
	if err != nil {
		return err
	}

	result, err := db.Query("SELECT id, red_champion_id, blue_champion_id, bets_red, bets_blue, winner, match_id FROM Fights ORDER BY match_id ASC")
	if err != nil {
		return err
	}
	fsql := "SELECT id, elo, wins, losses, total_bets FROM Champions WHERE id=$1"

	for result.Next() {
		m := &Match{}
		red, blue := &Fighter{}, &Fighter{}

		result.Scan(&m.Id, &m.RedId, &m.BlueId, &m.RedBets, &m.BlueBets, &m.Winner, &m.MatchId)
		fmt.Printf("Processing match #%05d\r", m.MatchId)
		if err := db.QueryRow(fsql, m.RedId).Scan(&red.Id, &red.Elo, &red.Win, &red.Loss, &red.TotalBets); err != nil {
			return err
		}
		if err := db.QueryRow(fsql, m.BlueId).Scan(&blue.Id, &blue.Elo, &blue.Win, &blue.Loss, &blue.TotalBets); err != nil {
			return err
		}

		red.TotalBets += m.RedBets
		blue.TotalBets += m.BlueBets
		UpdateFighterElo(red, blue, FightWinner(m.Winner))

		if err := r.UpdateFighter(red); err != nil {
			return err
		}
		if err := r.UpdateFighter(blue); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) StartTransaction() (*sql.Tx, error) {
	if r.single_conn {
		db, _ := r.open()
		return db.Begin()
	} else {
		return nil, errors.New("Cannot start transaction unless repository was created with OpenDb.")
	}
}

func (r *Repository) Close() {
	if r.single_conn {
		r.db.Close()
	}
}

func (r *Repository) open() (*sql.DB, error) {
	if r.single_conn {
		return r.db, nil
	} else {
		return sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s", r.user, r.password, r.dbname))
	}
}

func (r *Repository) close(db *sql.DB) {
	if !r.single_conn {
		db.Close()
	}
}
