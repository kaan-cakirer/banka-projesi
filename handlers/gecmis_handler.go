package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// IslemGecmisi: POST /gecmis (JSON Body: {"id": 1, "pin": "1234"})
func IslemGecmisi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var istek models.GecmisIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		fmt.Fprintf(w, "Geçersiz JSON verisi: %v", err)
		return
	}

	// 1. PIN Kontrolü
	var gercekPin string
	err = config.DB.QueryRow("SELECT pin FROM hesaplar WHERE id = ?", istek.ID).Scan(&gercekPin)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(w, "Hesap bulunamadı.")
			return
		}
		fmt.Fprintf(w, "Veritabanı hatası: %v", err)
		return
	}

	if istek.Pin != gercekPin {
		fmt.Fprintf(w, "Hatalı PIN Kodu! İşlem geçmişi görüntülenemez.")
		return
	}

	// 2. Son 10 İşlemi Çekme
	sorgu := `
	SELECT id, gonderen_id, alici_id, miktar, islem_tipi, tarih 
	FROM islemler 
	WHERE gonderen_id = ? OR alici_id = ? 
	ORDER BY id DESC 
	LIMIT 10`

	satirlar, err := config.DB.Query(sorgu, istek.ID, istek.ID)
	if err != nil {
		fmt.Fprintf(w, "İşlem geçmişi alınırken hata oluştu: %v", err)
		return
	}
	defer satirlar.Close()

	fmt.Fprintf(w, "=== HESAP İŞLEM GEÇMİŞİ (DEKONT) ===\n\n")

	islemSayaci := 0
	kullaniciIDStr := fmt.Sprintf("%d", istek.ID)

	for satirlar.Next() {
		var islem models.Islem
		var miktarKurus int

		err := satirlar.Scan(&islem.ID, &islem.GonderenID, &islem.AliciID, &miktarKurus, &islem.IslemTipi, &islem.Tarih)
		if err != nil {
			fmt.Fprintf(w, "Veri okunurken hata oluştu: %v", err)
			return
		}

		islem.MiktarTL = float64(miktarKurus) / 100.0
		islemSayaci++

		if islem.IslemTipi == "YATIRMA" {
			fmt.Fprintf(w, "[%s] PARA YATIRMA: +%.2f TL | Tarih: %s\n", islem.IslemTipi, islem.MiktarTL, islem.Tarih)
		} else if fmt.Sprintf("%d", islem.GonderenID) == kullaniciIDStr {
			fmt.Fprintf(w, "[GİDEN TRANSFER] Alıcı ID: %d | Miktar: -%.2f TL | Tarih: %s\n", islem.AliciID, islem.MiktarTL, islem.Tarih)
		} else {
			fmt.Fprintf(w, "[GELEN TRANSFER] Gönderen ID: %d | Miktar: +%.2f TL | Tarih: %s\n", islem.GonderenID, islem.MiktarTL, islem.Tarih)
		}
	}

	if islemSayaci == 0 {
		fmt.Fprintf(w, "Henüz hiç işlem geçmişiniz bulunmamaktadır.")
	}
}
