package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"banka-projesi/utils"
	"database/sql"
	"encoding/json"
	"net/http"
)

// BakiyeSorgula: POST /bakiye (JSON Body: {"id": 1, "pin": "1234"})
func BakiyeSorgula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.BakiyeIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	var isim string
	var bakiyeKurus int
	var gercekPin string

	sorgu := "SELECT isim, bakiye, pin FROM hesaplar WHERE id = ?"
	err = config.DB.QueryRow(sorgu, istek.ID).Scan(&isim, &bakiyeKurus, &gercekPin)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if istek.Pin != gercekPin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN Kodu! Bakiye görüntülenemez", nil)
		return
	}

	bakiyeTL := float64(bakiyeKurus) / 100.0

	// data içinde düzenli bir struct/map dönüyoruz
	hesapVerisi := map[string]any{
		"id":     istek.ID,
		"isim":   isim,
		"bakiye": bakiyeTL,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Bakiye başarıyla sorgulandı", hesapVerisi)
}

// HesapAc: POST /hesap-ac (JSON Body: {"isim": "Kaan", "bakiye": 5000, "pin": "9999"})
func HesapAc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.HesapAcIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	if istek.Isim == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Lütfen geçerli bir isim giriniz", nil)
		return
	}

	if istek.Pin == "" {
		istek.Pin = "1234"
	}

	baslangicKurus := int(istek.Bakiye * 100)

	sorgu := "INSERT INTO hesaplar (isim, bakiye, pin) VALUES (?, ?, ?)"
	sonuc, err := config.DB.Exec(sorgu, istek.Isim, baslangicKurus, istek.Pin)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Hesap oluşturulurken hata oluştu", nil)
		return
	}

	yeniID, _ := sonuc.LastInsertId()

	yeniHesap := models.Hesap{
		ID:     int(yeniID),
		Isim:   istek.Isim,
		Bakiye: istek.Bakiye,
	}

	utils.JSONResponse(w, http.StatusCreated, true, "Hesap başarıyla oluşturuldu", yeniHesap)
}

// ParaYatir: POST /para-yatir (JSON Body: {"id": 1, "miktar": 500})
func ParaYatir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.ParaYatirIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON veya miktar verisi", nil)
		return
	}

	yatirilacakKurus := int(istek.Miktar * 100)

	// Bakiye artırma işlemi
	sonuc, err := config.DB.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", yatirilacakKurus, istek.ID)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Para yatırılırken hata oluştu", nil)
		return
	}

	etkilenenSatir, _ := sonuc.RowsAffected()
	if etkilenenSatir == 0 {
		utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı! ID numarasını kontrol ediniz", nil)
		return
	}

	// İşlem geçmişine (islemler) ekleme
	config.DB.Exec("INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)", istek.ID, istek.ID, yatirilacakKurus, "YATIRMA")

	islemDetay := map[string]any{
		"hesap_id":         istek.ID,
		"yatirilan_miktar": istek.Miktar,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Para yatırma işlemi başarılı", islemDetay)
}
