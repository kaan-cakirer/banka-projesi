package models

// DB'den okuma yaparken ve yanıt dönerken kullanacağımız ana hesap modeli
type Hesap struct {
	ID     int     `json:"id"`
	Isim   string  `json:"isim"`
	Bakiye float64 `json:"bakiye"`
	Pin    string  `json:"pin,omitempty"` // omitempty: JSON yanıtlarında PIN görünmesin diye
}

// POST /bakiye için gelen JSON gövdesi
type BakiyeIstegi struct {
	ID  int    `json:"id"`
	Pin string `json:"pin"`
}

// POST /hesap-ac için gelen JSON gövdesi
type HesapAcIstegi struct {
	Isim   string  `json:"isim"`
	Bakiye float64 `json:"bakiye"`
	Pin    string  `json:"pin"`
}

// POST /para-yatir için gelen JSON gövdesi
type ParaYatirIstegi struct {
	ID     int     `json:"id"`
	Miktar float64 `json:"miktar"`
}
