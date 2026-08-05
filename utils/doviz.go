package utils

import (
	"banka-projesi/models"
	"encoding/json"
	"fmt"
	"net/http"
)

func KurGetir(dovizKodu string) (float64, error) {

	cevap, err := http.Get("https://open.er-api.com/v6/latest/TRY")
	if err != nil {
		return 0, fmt.Errorf("Döviz Servisine Ulaşırken Hata Oluştu: %v", err)
	}
	defer cevap.Body.Close()

	var veri models.DovizResponse
	err = json.NewDecoder(cevap.Body).Decode(&veri)

	if err != nil {
		return 0, fmt.Errorf("Döviz verisi işlenemedi: %v", err)
	}

	oran, varmi := veri.Rates[dovizKodu]
	if !varmi {
		return 0, fmt.Errorf("Geçersiz döviz kodu: %s", dovizKodu)
	}
	tlKarsiliği := 1.00 / oran

	return tlKarsiliği, nil
}
