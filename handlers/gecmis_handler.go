package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"banka-projesi/utils"
	"database/sql"
	"encoding/json"
	"net/http"
)

// IslemGecmisi: POST /gecmis (JSON Body: {"id": 1, "pin": "1234"})
func IslemGecmisi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.GecmisIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	// 1. PIN Kontrolü
	var gercekPin string
	err = config.DB.QueryRow("SELECT pin FROM hesaplar WHERE id = ?", istek.ID).Scan(&gercekPin)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if istek.Pin != gercekPin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN Kodu! İşlem geçmişi görüntülenemez", nil)
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
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem geçmişi alınırken hata oluştu", nil)
		return
	}
	defer satirlar.Close()

	islemler := make([]models.Islem, 0)

	for satirlar.Next() {
		var islem models.Islem
		var miktarKurus int

		err := satirlar.Scan(&islem.ID, &islem.GonderenID, &islem.AliciID, &miktarKurus, &islem.IslemTipi, &islem.Tarih)
		if err != nil {
			utils.JSONResponse(w, http.StatusInternalServerError, false, "Veri okunurken hata oluştu", nil)
			return
		}

		islem.MiktarTL = float64(miktarKurus) / 100.0
		islemler = append(islemler, islem)
	}

	if err = satirlar.Err(); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem geçmişi okunurken hata oluştu", nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "İşlem geçmişi başarıyla getirildi", islemler)
}
