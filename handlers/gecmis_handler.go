package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"banka-projesi/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

// IslemGecmisi: POST /gecmis (JSON Body: {"id": 1, "pin": "1234"})
func IslemGecmisi(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 DEBUG: /gecmis isteği sunucuya ulaştı.")

	if r.Method != http.MethodPost {
		log.Println("❌ DEBUG HATA: İstek metodu POST değil:", r.Method)
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.GecmisIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		log.Println("❌ DEBUG HATA: Gelen JSON okunamadı:", err)
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	log.Printf("🔍 DEBUG: Okunan İstek -> Hesap ID: %d, PIN: %s\n", istek.ID, istek.Pin)

	// 1. PIN Kontrolü
	var gercekPin string
	err = config.DB.QueryRow("SELECT pin FROM hesaplar WHERE id = ?", istek.ID).Scan(&gercekPin)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("❌ DEBUG HATA: Veritabanında bu ID bulunamadı:", istek.ID)
			utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı", nil)
			return
		}
		log.Println("❌ DEBUG HATA: PIN sorgulanırken DB hatası:", err)
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if istek.Pin != gercekPin {
		log.Printf("❌ DEBUG HATA: PIN eşleşmedi! Gelen: %s, Gerçek: %s\n", istek.Pin, gercekPin)
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN Kodu! İşlem geçmişi görüntülenemez", nil)
		return
	}

	log.Println("✅ DEBUG: PIN doğrulandı. İşlem geçmişi sorgulanıyor...")

	// 2. Son 10 işlemi, normal ve döviz işlemlerini birlikte çek
	sorgu := `
	SELECT id, gonderen_id, alici_id, miktar, islem_tipi, tarih, NULL as doviz_kodu
	FROM islemler
	WHERE gonderen_id = ? OR alici_id = ?
	UNION ALL
	SELECT id, hesap_id as gonderen_id, NULL as alici_id, harcanan_tl_kurus as miktar, islem_tipi, tarih, doviz_kodu
	FROM doviz_islemleri
	WHERE hesap_id = ?
	ORDER BY tarih DESC
	LIMIT 10`

	satirlar, err := config.DB.Query(sorgu, istek.ID, istek.ID, istek.ID)
	if err != nil {
		log.Println("❌ DEBUG HATA: Islemler tablosu sorgulanamadı (Tablo var mı?):", err)
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem geçmişi alınırken hata oluştu", nil)
		return
	}
	defer satirlar.Close()

	islemler := make([]models.Islem, 0)

	for satirlar.Next() {
		var islem models.Islem
		var miktarKurus int
		var aliciID sql.NullInt64
		var dovizKodu sql.NullString

		err := satirlar.Scan(
			&islem.ID,
			&islem.GonderenID,
			&aliciID,
			&miktarKurus,
			&islem.IslemTipi,
			&islem.Tarih,
			&dovizKodu,
		)
		if err != nil {
			log.Println("❌ DEBUG HATA: Satır Scan edilirken hata (Sütun tipleri uyuşuyor mu?):", err)
			utils.JSONResponse(w, http.StatusInternalServerError, false, "Veri okunurken hata oluştu", nil)
			return
		}

		if aliciID.Valid {
			islem.AliciID = int(aliciID.Int64)
		} else {
			islem.AliciID = 0
		}

		if dovizKodu.Valid {
			islem.DovizKodu = dovizKodu.String
		} else {
			islem.DovizKodu = ""
		}

		islem.MiktarTL = float64(miktarKurus) / 100.0
		islemler = append(islemler, islem)
	}

	if err = satirlar.Err(); err != nil {
		log.Println("❌ DEBUG HATA: Satır döngüsünde hata:", err)
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem geçmişi okunurken hata oluştu", nil)
		return
	}

	log.Printf("✅ DEBUG: Başarılı! Toplam %d adet işlem geçmişi bulundu ve JSON olarak dönülüyor.\n", len(islemler))
	utils.JSONResponse(w, http.StatusOK, true, "İşlem geçmişi başarıyla getirildi", islemler)
}
