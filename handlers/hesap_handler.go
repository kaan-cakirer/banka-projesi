package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"banka-projesi/config"
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
// Yeni bir usecase hatası eklediğinde buraya bir case eklemen yeterli.
func statusFromError(err error) int {
	switch {
	case errors.Is(err, usecase.ErrGecersizIsim),
		errors.Is(err, usecase.ErrNegatifBakiye):
		return http.StatusBadRequest
	case errors.Is(err, usecase.ErrHataliPin):
		return http.StatusUnauthorized
	case errors.Is(err, usecase.ErrHesapBulunamadi):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
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
	sonuc, err := config.DB.Exec("UPDATE hesaplar SET bakiye = bakiye + $1 WHERE id = $2", yatirilacakKurus, istek.ID)
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
	config.DB.Exec("INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES ($1, $2, $3, $4)", istek.ID, istek.ID, yatirilacakKurus, "YATIRMA")

	islemDetay := map[string]any{
		"hesap_id":         istek.ID,
		"yatirilan_miktar": istek.Miktar,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Para yatırma işlemi başarılı", islemDetay)
}
