package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"banka-projesi/models"
	"banka-projesi/usecase"
	"banka-projesi/utils"
)

// HesapHandler: HTTP katmanı. SQL bilmez, iş kuralı bilmez - sadece
// JSON'u parse edip usecase'e iletir, dönen sonucu/hatayı JSON'a çevirir.
type HesapHandler struct {
	usecase usecase.HesapUsecase
}

// NewHesapHandler: constructor. main.go'da usecase'i buraya vereceğiz.
func NewHesapHandler(u usecase.HesapUsecase) *HesapHandler {
	return &HesapHandler{usecase: u}
}

// HesapAc: POST /api/hesap-ac
func (h *HesapHandler) HesapAc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.HesapAcIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	hesap, err := h.usecase.HesapAc(istek.Isim, istek.Bakiye, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, "Hesap başarıyla oluşturuldu", hesap)
}

// BakiyeSorgula: POST /api/bakiye
func (h *HesapHandler) BakiyeSorgula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.BakiyeIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	hesap, err := h.usecase.BakiyeSorgula(istek.ID, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Bakiye başarıyla sorgulandı", hesap)
}

// statusFromError: usecase'in sentinel error'larını HTTP status koduna çevirir.
func statusFromError(err error) int {
	switch {
	// 1. Kullanıcının eksik/yanlış veri girdiği durumlar (400 Bad Request)
	case errors.Is(err, usecase.ErrGecersizIsim),
		errors.Is(err, usecase.ErrNegatifBakiye),
		errors.Is(err, usecase.ErrGecersizMiktar),
		errors.Is(err, usecase.ErrGecersizDovizKodu),
		errors.Is(err, usecase.ErrYetersizDoviz):
		return http.StatusBadRequest

	// 2. PIN Yanlışsa (401 Unauthorized)
	case errors.Is(err, usecase.ErrHataliPin):
		return http.StatusUnauthorized

	// 3. Hesap DB'de yoksa (404 Not Found)
	case errors.Is(err, usecase.ErrHesapBulunamadi),
		errors.Is(err, usecase.ErrGonderenBulunamadı),
		errors.Is(err, usecase.ErrAliciBulunamadı):
		return http.StatusNotFound

	// 3b. Yetersiz bakiye (400 Bad Request)
	case errors.Is(err, usecase.ErrYetersizBakiye):
		return http.StatusBadRequest

	// 4. Veritabanı çökmesi, işlem başarısızlığı gibi sistem hataları (500 Internal Server Error)
	// ErrBasarisizIslem, ErrOlusturmaHatasi, ErrGecmisAlinamadı gibi tanımladığınız
	// diğer tüm hatalar otomatik olarak bu 'default' bloğuna düşüp 500 dönecektir.
	default:
		return http.StatusInternalServerError
	}
}

// ParaYatir: POST /para-yatir (JSON Body: {"id": 1, "miktar": 500, "pin": "1234"})
func (h *HesapHandler) ParaYatir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.MiktarIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	yatirma, err := h.usecase.ParaYatir(istek.ID, istek.Miktar, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Para yatırma işlemi başarılı", yatirma)
}

func (h *HesapHandler) ParaCek(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.MiktarIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	cekme, err := h.usecase.ParaCek(istek.ID, istek.Miktar, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Para çekme işlemi başarılı", cekme)
}

func (h *HesapHandler) ParaGonder(w http.ResponseWriter, r *http.Request) {
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

	transfer, err := h.usecase.Transfer(istek.GonderenID, istek.AliciID, istek.Miktar, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Transfer işlemi başarılı", transfer)
}

// IslemGecmisi: POST /api/islem-gecmisi
func (h *HesapHandler) IslemGecmisi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.GecmisIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	// Usecase'i çağırıyoruz
	islemler, err := h.usecase.IslemSorgula(istek.ID, istek.Pin)
	if err != nil {
		// Zaten yazdığınız harika yardımcı fonksiyonu kullanıyoruz!
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "İşlem geçmişi başarıyla getirildi", islemler)
}
