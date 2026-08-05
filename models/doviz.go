package models

type DovizResponse struct {
	BaseCode string             `json:"base_code"`
	Rates    map[string]float64 `json:"rates"`
}

type Varlik struct {
	ID        int     `json:"id"`
	HesapID   int     `json:"hesap_id"`
	DovizKodu string  `json:"doviz_kodu"`
	Miktar    float64 `json:"miktar"`
}

type DovizIslemIstegi struct {
	HesapID   int     `json:"hesap_id"`
	DovizKodu string  `json:"doviz_kodu"`
	Miktar    float64 `json:"miktar"`
	Pin       string  `json:"pin"`
}

type DovizIslem struct {
	ID         int     `json:"id"`
	HesapID    int     `json:"hesap_id"`
	DovizKodu  string  `json:"doviz_kodu"`
	Miktar     float64 `json:"miktar"`
	HarcananTL float64 `json:"harcanan_tl"`
	KurFiyati  float64 `json:"kur_fiyati"`
	IslemTipi  string  `json:"islem_tipi"`
	Tarih      string  `json:"tarih"`
}
