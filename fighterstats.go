package spicerack

type FighterStats struct {
	RedName  string `json:"p1name"`
	RedTier  string `json:"p1tier"`
	RedLife  int    `json:"p1life,string"`
	RedMeter int    `json:"p1meter,string"`

	BlueName  string `json:"p2name"`
	BlueTier  string `json:"p2tier"`
	BlueLife  int    `json:"p2life,string"`
	BlueMeter int    `json:"p2meter,string"`
}
