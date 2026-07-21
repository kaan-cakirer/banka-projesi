package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// BakiyeSorgula: POST /bakiye (JSON Body: {"id": 1, "pin": "1234"})
func BakiyeSorgula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var istek models.BakiyeIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		fmt.Fprintf(w, "Geçersiz JSON verisi: %v", err)
		return
	}

	var isim string
	var bakiyeKurus int
	var gercekPin string

	sorgu := "SELECT isim, bakiye, pin FROM hesaplar WHERE id = ?"
	err = config.DB.QueryRow(sorgu, istek.ID).Scan(&isim, &bakiyeKurus, &gercekPin)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(w, "Hesap bulunamadı.")
			return
		}
		fmt.Fprintf(w, "Veritabanı hatası: %v", err)
		return
	}

	if istek.Pin != gercekPin {
		fmt.Fprintf(w, "Hatalı PIN Kodu! Bakiye görüntülenemez.")
		return
	}

	bakiyeTL := float64(bakiyeKurus) / 100.0
	fmt.Fprintf(w, "Hesap Sahibi: %s\nBakiye: %.2f TL", isim, bakiyeTL)
}

// HesapAc: POST /hesap-ac (JSON Body: {"isim": "Kaan", "bakiye": 5000, "pin": "9999"})
func HesapAc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var istek models.HesapAcIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil {
		fmt.Fprintf(w, "Geçersiz JSON verisi: %v", err)
		return
	}

	if istek.Isim == "" {
		fmt.Fprintf(w, "Lütfen geçerli bir isim giriniz.")
		return
	}

	if istek.Pin == "" {
		istek.Pin = "1234"
	}

	baslangicKurus := int(istek.Bakiye * 100)

	sorgu := "INSERT INTO hesaplar (isim, bakiye, pin) VALUES (?, ?, ?)"
	sonuc, err := config.DB.Exec(sorgu, istek.Isim, baslangicKurus, istek.Pin)
	if err != nil {
		fmt.Fprintf(w, "Hesap oluşturulurken hata oluştu: %v", err)
		return
	}

	yeniID, _ := sonuc.LastInsertId()
	fmt.Fprintf(w, "Hesap başarıyla oluşturuldu!\nHesap ID: %d\nSahibi: %s\nPIN: %s", yeniID, istek.Isim, istek.Pin)
}

// ParaYatir: POST /para-yatir (JSON Body: {"id": 1, "miktar": 500})
func ParaYatir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var istek models.ParaYatirIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		fmt.Fprintf(w, "Geçersiz JSON veya miktar verisi")
		return
	}

	yatirilacakKurus := int(istek.Miktar * 100)

	// Bakiye artırma işlemi
	sonuc, err := config.DB.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", yatirilacakKurus, istek.ID)
	if err != nil {
		fmt.Fprintf(w, "Para yatırılırken hata oluştu: %v", err)
		return
	}

	etkilenenSatir, _ := sonuc.RowsAffected()
	if etkilenenSatir == 0 {
		fmt.Fprintf(w, "Hesap bulunamadı! ID numarasını kontrol ediniz.")
		return
	}

	// İşlem geçmişine (islemler) ekleme
	config.DB.Exec("INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)", istek.ID, istek.ID, yatirilacakKurus, "YATIRMA")

	fmt.Fprintf(w, "Para yatırma başarılı! %.2f TL yatırıldı.", istek.Miktar)
}
