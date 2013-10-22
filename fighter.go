package spicerack

import "time"

type Fighter struct {
	Id        int
	Win       int
	Loss      int
	Elo       int
	Name      string
	TotalBets int
	Created   time.Time
	Updated   time.Time
}

func (f *Fighter) TotalMatches() int {
	return f.Win + f.Loss
}

func (f *Fighter) WinRate() float32 {
	return (float32(f.Win) / float32(f.TotalMatches())) * 100
}
