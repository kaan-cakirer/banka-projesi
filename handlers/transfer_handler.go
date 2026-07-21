package handlers

import (
	"banka-projesi/config"
	"fmt"
	"net/http"
	"strconv"
)

func ParaGonder(w http.ResponseWriter, r *http.Request) {
	gonderenID := r.URL.Query().Get("gonderen")
	alanID := r.URL.Query().Get("alan")
	miktarStr := r.URL.Query().Get("miktar")
	pin := r.URL.Query().Get("pin")

	miktarSayi, err := strconv.ParseFloat(miktarStr, 64)
	if err != nil || miktarSayi <= 0 {
		fmt.Fprintf(w, "Geçerli bir miktar giriniz")
		return
	}
	gonderilecekKurus := int(miktarSayi * 100)

	var gonderenBakiye int
	var gercekPin string

	sorgu := "SELECT bakiye, pin FROM hesaplar WHERE id = ?"
	err = config.DB.QueryRow(sorgu, gonderenID).Scan(&gonderenBakiye, &gercekPin)

	if err != nil {
		fmt.Fprintf(w, "Gönderen hesap bulunamadı")
		return
	}

	// PIN Kontrolü (String kıyası)
	if gercekPin != pin {
		fmt.Fprintf(w, "Hatalı PIN kodu! İşlem reddedildi.")
		return
	}

	if gonderenBakiye < gonderilecekKurus {
		fmt.Fprintf(w, "Yetersiz Bakiye!")
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		fmt.Fprintf(w, "İşlem Başlatılamadı")
		return
	}

	// 1. Gönderenden düş
	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye - ? WHERE id = ?", gonderilecekKurus, gonderenID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(w, "Gönderilen Bakiye Düşülürken Bir Hata Oluştu: %v", err)
		return
	}

	// 2. Alıcıya ekle
	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", gonderilecekKurus, alanID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(w, "Gönderilen Bakiye Yüklenirken Bir Hata Oluştu: %v", err)
		return
	}

	// 3. İşlem Geçmişine Kaydet
	islemSorgusu := "INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(islemSorgusu, gonderenID, alanID, gonderilecekKurus, "TRANSFER")
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(w, "İşlem geçmişi kaydedilirken hata oluştu: %v", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(w, "İşlem Onaylanmadı: %v", err)
		return
	}

	fmt.Fprintf(w, "Transfer Başarılı!\n%.2f TL Tutarındaki Miktar %s Kullanıcısından %s Kullanıcısına Başarıyla Aktarıldı.", miktarSayi, gonderenID, alanID)
}
