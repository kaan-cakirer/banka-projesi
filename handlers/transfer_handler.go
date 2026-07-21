package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"banka-projesi/utils"
	"database/sql"
	"encoding/json"
	"net/http"
)

// ParaGonder: POST /para-gonder (JSON Body: {"gonderen_id": 1, "alici_id": 2, "miktar": 100, "pin": "1234"})
func ParaGonder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.TransferIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi veya miktar", nil)
		return
	}

	gonderilecekKurus := int(istek.Miktar * 100)

	var gonderenBakiye int
	var gercekPin string

	sorgu := "SELECT bakiye, pin FROM hesaplar WHERE id = ?"
	err = config.DB.QueryRow(sorgu, istek.GonderenID).Scan(&gonderenBakiye, &gercekPin)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusNotFound, false, "Gönderen hesap bulunamadı", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	// PIN Kontrolü
	if gercekPin != istek.Pin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN kodu! İşlem reddedildi", nil)
		return
	}

	if gonderenBakiye < gonderilecekKurus {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Yetersiz bakiye", nil)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem başlatılamadı", nil)
		return
	}

	// 1. Gönderenin bakiyesini düş
	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye - ? WHERE id = ?", gonderilecekKurus, istek.GonderenID)
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Gönderilen bakiye düşülürken hata oluştu", nil)
		return
	}

	// 2. Alıcının bakiyesini artır
	sonuc, err := tx.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", gonderilecekKurus, istek.AliciID)
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Gönderilen bakiye yüklenirken hata oluştu", nil)
		return
	}

	etkilenenSatir, _ := sonuc.RowsAffected()
	if etkilenenSatir == 0 {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusNotFound, false, "Alıcı hesap bulunamadı", nil)
		return
	}

	// 3. İşlem geçmişine kaydet
	islemSorgusu := "INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(islemSorgusu, istek.GonderenID, istek.AliciID, gonderilecekKurus, "TRANSFER")
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem geçmişi kaydedilirken hata oluştu", nil)
		return
	}

	err = tx.Commit()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem onaylanamadı", nil)
		return
	}

	transferDetay := map[string]any{
		"gonderen_id": istek.GonderenID,
		"alici_id":    istek.AliciID,
		"miktar":      istek.Miktar,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Transfer işlemi başarıyla tamamlandı", transferDetay)
}
