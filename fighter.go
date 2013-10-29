package spicerack

import (
	"fmt"
	"time"
)

const (
	stats_format       string = "(%s) E:%s W\x02\x0309%d\x03\x02/L\x02\x0304%d\x03\x02 (\x02%02.1f%%\x02)"
	elo_format         string = "\x03%02d%d\x03"
	godlike_elo_format string = "\x1F\x02\x0309-%d-\x03\x02\x1F"
	tier_format        string = "\x02\x03%02d%s\x03\x02"
)

type Fighter struct {
	Id          int
	Win         int
	Loss        int
	Elo         int
	Name        string
	TotalBets   int
	CharacterId int
	Tier        int
	Created     time.Time
	Updated     time.Time
}

func (f *Fighter) TotalMatches() int {
	return f.Win + f.Loss
}

func (f *Fighter) WinRate() float32 {
	tm := f.TotalMatches()
	if tm == 0 {
		return 0.0
	}
	return (float32(f.Win) / float32(tm)) * 100
}

func (f *Fighter) IrcTierFormat() string {
	switch f.Tier {
	case TIER_S: // Green
		return fmt.Sprintf(tier_format, 9, "S")
	case TIER_A: // Red
		return fmt.Sprintf(tier_format, 5, "A")
	case TIER_B: // Blue
		return fmt.Sprintf(tier_format, 12, "B")
	case TIER_P: // Brown
		return fmt.Sprintf(tier_format, 7, "P")
	}

	return fmt.Sprintf(tier_format, 13, "?")
}

func (f *Fighter) IrcEloFormat() string {
	color := 0

	if f.Elo < 400 {
		color = 4 // Red
	} else if f.Elo >= 400 && f.Elo < 700 {
		color = 7 // Orange
	} else if f.Elo >= 700 && f.Elo < 1000 {
		color = 3 // Dark Green
	} else if f.Elo >= 1000 && f.Elo < 1300 {
		color = 9 // Light Green
	} else if f.Elo >= 1300 {
		return fmt.Sprintf(godlike_elo_format, f.Elo)
	}

	return fmt.Sprintf(elo_format, color, f.Elo)
}

func (f *Fighter) IrcStats() string {
	return fmt.Sprintf(stats_format, f.IrcTierFormat(), f.IrcEloFormat(), f.Win, f.Loss, f.WinRate())
}
