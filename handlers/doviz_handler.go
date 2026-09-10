package handlers

import (
	"encoding/json"
	"net/http"

	"banka-projesi/models"
	"banka-projesi/usecase"
	"banka-projesi/utils"
)

// DovizHandler: HTTP katmanı. SQL bilmez, iş kuralı bilmez - sadece
// JSON'u parse edip usecase'e iletir, dönen sonucu/hatayı JSON'a çevirir.
type DovizHandler struct {
	usecase usecase.DovizUsecase
}

// NewDovizHandler: constructor. main.go'da usecase'i buraya vereceğiz.
func NewDovizHandler(u usecase.DovizUsecase) *DovizHandler {
	return &DovizHandler{usecase: u}
}

// DovizAl: POST /api/doviz-al
func (h *DovizHandler) DovizAl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.DovizIslemIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi veya miktar", nil)
		return
	}

	islem, err := h.usecase.DovizAl(istek.HesapID, istek.DovizKodu, istek.Miktar, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Döviz başarıyla satın alındı", islem)
}

// DovizSat: POST /api/doviz-sat
func (h *DovizHandler) DovizSat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.DovizIslemIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi veya miktar", nil)
		return
	}

	islem, err := h.usecase.DovizSat(istek.HesapID, istek.DovizKodu, istek.Miktar, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Döviz başarıyla satıldı", islem)
}

// DovizKur: POST /api/doviz-kur (JSON Body: {"doviz_kodu": "USD"})
// İşlem yapmaz - sadece seçili dövizin güncel TL karşılığını döner.
func (h *DovizHandler) DovizKur(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.DovizKurIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	kur, err := h.usecase.KurGetir(istek.DovizKodu)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Kur bilgisi getirildi", kur)
}

// VarliklariGetir: POST /api/varliklar (JSON Body: {"id": 1, "pin": "1234"})
func (h *DovizHandler) VarliklariGetir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.GecmisIstegi
	if err := json.NewDecoder(r.Body).Decode(&istek); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	varliklar, err := h.usecase.VarliklarGetir(istek.ID, istek.Pin)
	if err != nil {
		status := statusFromError(err)
		utils.JSONResponse(w, status, false, err.Error(), nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Varlıklar başarıyla getirildi", varliklar)
}
