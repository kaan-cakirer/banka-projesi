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

func DovizAl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("❌ DEBUG HATA: İstek metodu POST değil:", r.Method)
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.DovizIslemIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi veya miktar", nil)
		return
	}

	kurOran, err := utils.KurGetir(istek.DovizKodu)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz döviz kodu veya kur sunucusuna ulaşılamadı", nil)
		return
	}

	dovizCent := int(istek.Miktar * 100)

	toplamTL := istek.Miktar * kurOran
	harcanacakTLKurus := int(toplamTL * 100)

	var bakiye int
	var gercekPin string

	sorgu := "SELECT bakiye, pin FROM hesaplar WHERE id = $1"
	err = config.DB.QueryRow(sorgu, istek.HesapID).Scan(&bakiye, &gercekPin)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if gercekPin != istek.Pin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN kodu! İşlem reddedildi", nil)
		return
	}

	if bakiye < harcanacakTLKurus {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Yetersiz TL bakiyesi", nil)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem başlatılamadı", nil)
		return
	}

	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye - $1 WHERE id = $2", harcanacakTLKurus, istek.HesapID)
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "TL bakiyesi düşülürken hata oluştu", nil)
		return
	}

	varlikSorgusu := `
	INSERT INTO varliklar (hesap_id, doviz_kodu, miktar) 
	VALUES ($1, $2, $3) 
	ON CONFLICT(hesap_id, doviz_kodu) 
	DO UPDATE SET miktar = miktar + EXCLUDED.miktar`

	_, err = tx.Exec(varlikSorgusu, istek.HesapID, istek.DovizKodu, dovizCent)
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Döviz varlığı güncellenirken hata oluştu", nil)
		return
	}
	dekontSorgusu := `
	INSERT INTO doviz_islemleri (hesap_id, doviz_kodu, miktar_cent, harcanan_tl_kurus, kur_fiyati, islem_tipi) 
	VALUES ($1, $2, $3, $4, $5, 'ALIM')`

	_, err = tx.Exec(dekontSorgusu, istek.HesapID, istek.DovizKodu, dovizCent, harcanacakTLKurus, kurOran)
	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem kaydı oluşturulurken hata oluştu", nil)
		return
	}
	err = tx.Commit()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem tamamlanırken veritabanı hatası oluştu", nil)
		return
	}
	kalanTL := float64(bakiye-harcanacakTLKurus) / 100.0

	yanit := map[string]interface{}{
		"mesaj":      "Döviz alım işlemi başarılı",
		"satilan":    istek.Miktar,
		"doviz_kodu": istek.DovizKodu,
		"kalan_tl":   kalanTL,
		"kur":        kurOran,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Döviz başarıyla satın alındı", yanit)
}

func DovizSat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("❌ DEBUG HATA: İstek metodu POST değil:", r.Method)
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.DovizIslemIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi veya miktar", nil)
		return
	}

	kurOran, err := utils.KurGetir(istek.DovizKodu)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz döviz kodu veya kur sunucusuna ulaşılamadı", nil)
		return
	}

	dovizCent := int(istek.Miktar * 100)
	kazanilanTLKurus := int(istek.Miktar * kurOran * 100)

	var bakiye int
	var gercekPin string

	err = config.DB.QueryRow(
		"SELECT bakiye, pin FROM hesaplar WHERE id = $1",
		istek.HesapID,
	).Scan(&bakiye, &gercekPin)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusNotFound, false, "Hesap bulunamadı", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if gercekPin != istek.Pin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN kodu! İşlem reddedildi", nil)
		return
	}

	// Döviz bakiyesi kontrolü
	var mevcutDoviz int

	err = config.DB.QueryRow(
		"SELECT miktar FROM varliklar WHERE hesap_id = $1 AND doviz_kodu = $2",
		istek.HesapID,
		istek.DovizKodu,
	).Scan(&mevcutDoviz)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONResponse(w, http.StatusBadRequest, false, "Bu döviz hesabınızda bulunmuyor", nil)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Veritabanı hatası", nil)
		return
	}

	if mevcutDoviz < dovizCent {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Yetersiz döviz bakiyesi", nil)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem başlatılamadı", nil)
		return
	}

	// TL ekle
	_, err = tx.Exec(
		"UPDATE hesaplar SET bakiye = bakiye + $1 WHERE id = $2",
		kazanilanTLKurus,
		istek.HesapID,
	)

	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "TL bakiyesi güncellenemedi", nil)
		return
	}

	// Döviz düş
	_, err = tx.Exec(
		"UPDATE varliklar SET miktar = miktar - $1 WHERE hesap_id = $2 AND doviz_kodu = $3",
		dovizCent,
		istek.HesapID,
		istek.DovizKodu,
	)

	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Döviz varlığı güncellenemedi", nil)
		return
	}

	// Sıfır kaldıysa sil
	_, err = tx.Exec(
		"DELETE FROM varliklar WHERE hesap_id = $1 AND doviz_kodu = $2 AND miktar = 0",
		istek.HesapID,
		istek.DovizKodu,
	)

	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Varlık kaydı silinemedi", nil)
		return
	}

	// Dekont
	_, err = tx.Exec(
		`INSERT INTO doviz_islemleri
		(hesap_id, doviz_kodu, miktar_cent, harcanan_tl_kurus, kur_fiyati, islem_tipi)
		VALUES ($1, $2, $3, $4, $5, 'SATIM')`,
		istek.HesapID,
		istek.DovizKodu,
		dovizCent,
		kazanilanTLKurus,
		kurOran,
	)

	if err != nil {
		tx.Rollback()
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem kaydı oluşturulamadı", nil)
		return
	}

	err = tx.Commit()
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "İşlem tamamlanamadı", nil)
		return
	}

	kalanTL := float64(bakiye+kazanilanTLKurus) / 100.0

	yanit := map[string]interface{}{
		"mesaj":      "Döviz satım işlemi başarılı",
		"satilan":    istek.Miktar,
		"doviz_kodu": istek.DovizKodu,
		"kalan_tl":   kalanTL,
		"kur":        kurOran,
	}

	utils.JSONResponse(w, http.StatusOK, true, "Döviz başarıyla satıldı", yanit)
}

// VarliklariGetir: POST /api/varliklar (JSON Body: {"id": 1, "pin": "1234"})
func VarliklariGetir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.JSONResponse(w, http.StatusMethodNotAllowed, false, "Sadece POST istekleri kabul edilir", nil)
		return
	}

	var istek models.GecmisIstegi // id ve pin barındıran istek modeli
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, "Geçersiz JSON verisi", nil)
		return
	}

	// 1. PIN Doğrulaması
	var gercekPin string
	err = config.DB.QueryRow("SELECT pin FROM hesaplar WHERE id = $1", istek.ID).Scan(&gercekPin)
	if err != nil || gercekPin != istek.Pin {
		utils.JSONResponse(w, http.StatusUnauthorized, false, "Hatalı PIN veya hesap bulunamadı", nil)
		return
	}

	// 2. Kullanıcının Döviz Varlıklarını Sorgula
	rows, err := config.DB.Query("SELECT id, hesap_id, doviz_kodu, miktar FROM varliklar WHERE hesap_id = $1", istek.ID)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Varlıklar sorgulanırken hata oluştu", nil)
		return
	}
	defer rows.Close()

	varliklar := make([]models.Varlik, 0)
	for rows.Next() {
		var v models.Varlik
		var miktarCent int
		if err := rows.Scan(&v.ID, &v.HesapID, &v.DovizKodu, &miktarCent); err != nil {
			continue
		}
		v.Miktar = float64(miktarCent) / 100.0 // Cent -> Ana birim dönüşümü (Örn: 1050 -> 10.50)
		varliklar = append(varliklar, v)
	}
	if err = rows.Err(); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, "Varlıklar okunurken hata oluştu", nil)
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, "Varlıklar başarıyla getirildi", varliklar)
}
