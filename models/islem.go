package models

// DB'deki islemler tablosu karşılığı
type Islem struct {
	ID         int     `json:"id"`
	GonderenID int     `json:"gonderen_id"`
	AliciID    int     `json:"alici_id"`
	MiktarTL   float64 `json:"miktar_tl"`
	IslemTipi  string  `json:"islem_tipi"`
	Tarih      string  `json:"tarih"`
}

// POST /para-gonder için gelen JSON gövdesi
type TransferIstegi struct {
	GonderenID int     `json:"gonderen_id"`
	AliciID    int     `json:"alici_id"`
	Miktar     float64 `json:"miktar"`
	Pin        string  `json:"pin"`
}

// POST /gecmis için gelen JSON gövdesi
type GecmisIstegi struct {
	ID  int    `json:"id"`
	Pin string `json:"pin"`
}
