package models

type Islem struct {
	ID         int     `json:"id"`
	GonderenID int     `json:"gonderen_id"`
	AliciID    int     `json:"alici_id"`
	MiktarTL   float64 `json:"miktar_tl"`
	IslemTipi  string  `json:"islem_tipi"`
	Tarih      string  `json:"tarih"`
}
