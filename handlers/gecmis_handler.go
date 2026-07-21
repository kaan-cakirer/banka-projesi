package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"database/sql"
	"fmt"
	"net/http"
)

func IslemGecmisi(w http.ResponseWriter, r *http.Request) {
	kullaniciID := r.URL.Query().Get("id")
	gelenPin := r.URL.Query().Get("pin")

	var gercekPin string
	err := config.DB.QueryRow("SELECT pin FROM hesaplar WHERE id = ?", kullaniciID).Scan(&gercekPin)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(w, "Hesap bulunamadı.")
			return
		}
		fmt.Fprintf(w, "Veritabanı hatası: %v", err)
		return
	}

	if gelenPin != gercekPin {
		fmt.Fprintf(w, "Hatalı PIN Kodu! İşlem geçmişi görüntülenemez.")
		return
	}

	sorgu := `	
	SELECT id, gonderen_id, alici_id, miktar, islem_tipi, tarih 
	FROM islemler 
	WHERE gonderen_id = ? OR alici_id = ? 
	ORDER BY id DESC 
	LIMIT 10`

	satirlar, err := config.DB.Query(sorgu, kullaniciID, kullaniciID)

	if err != nil {
		fmt.Fprintf(w, "Kullanıcı Bilgileri Alınırken Bir Hata Oluştu: %v", err)
		return
	}
	defer satirlar.Close()

	fmt.Fprintf(w, "=== HESAP İŞLEM GEÇMİŞİ (DEKONT) ===\n\n")

	islemSayaci := 0

	for satirlar.Next() {
		var islem models.Islem
		var miktarKurus int

		err := satirlar.Scan(&islem.ID, &islem.GonderenID, &islem.AliciID, &miktarKurus, &islem.IslemTipi, &islem.Tarih)

		if err != nil {
			fmt.Fprintf(w, "Veri Okunurken Hata Oluştu: %v", err)
			return
		}
		islem.MiktarTL = float64(miktarKurus) / 100.0
		islemSayaci++

		if islem.IslemTipi == "YATIRMA" {
			fmt.Fprintf(w, "[%s] PARA YATIRMA: +%.2f TL | Tarih: %s\n", islem.IslemTipi, islem.MiktarTL, islem.Tarih)
		} else if islem.GonderenID == islem.AliciID {
			fmt.Fprintf(w, "[%s] Hesaba Ekleme: +%.2f TL | Tarih: %s\n", islem.IslemTipi, islem.MiktarTL, islem.Tarih)
		} else if fmt.Sprintf("%d", islem.GonderenID) == kullaniciID {
			fmt.Fprintf(w, "[GİDEN TRANSFER] Alıcı ID: %d | Miktar: -%.2f TL | Tarih: %s\n", islem.AliciID, islem.MiktarTL, islem.Tarih)
		} else {
			fmt.Fprintf(w, "[GELEN TRANSFER] Gönderen ID: %d | Miktar: +%.2f TL | Tarih: %s\n", islem.GonderenID, islem.MiktarTL, islem.Tarih)
		}

	}

	if islemSayaci == 0 {
		fmt.Fprintf(w, "Henüz hiç işlem geçmişiniz bulunmamaktadır.")
	}

	if err = satirlar.Err(); err != nil {
		fmt.Fprintf(w, "Satirlar okunurken hata olustu: %v", err)
		return
	}

}
