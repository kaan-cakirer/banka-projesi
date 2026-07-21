package handlers

import (
	"banka-projesi/config"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

// BakiyeSorgula: /bakiye?id=1&pin=1234
func BakiyeSorgula(w http.ResponseWriter, r *http.Request) {
	gelenID := r.URL.Query().Get("id")
	gelenPin := r.URL.Query().Get("pin")

	var isim string
	var bakiyeKurus int
	var gercekPin string

	sorgu := "SELECT isim, bakiye, pin FROM hesaplar WHERE id = ?"
	err := config.DB.QueryRow(sorgu, gelenID).Scan(&isim, &bakiyeKurus, &gercekPin)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(w, "Hesap bulunamadı.")
			return
		}
		fmt.Fprintf(w, "Veritabanı hatası: %v", err)
		return
	}

	if gelenPin != gercekPin {
		fmt.Fprintf(w, "Hatalı PIN Kodu! Bakiye görüntülenemez.")
		return
	}

	bakiyeTL := float64(bakiyeKurus) / 100.0
	fmt.Fprintf(w, "Hesap Sahibi: %s\nBakiye: %.2f TL", isim, bakiyeTL)
}

// HesapAc: /hesap-ac?isim=Ahmet&bakiye=5000&pin=4321
func HesapAc(w http.ResponseWriter, r *http.Request) {
	isim := r.URL.Query().Get("isim")
	bakiyeStr := r.URL.Query().Get("bakiye")
	pin := r.URL.Query().Get("pin")

	if isim == "" {
		fmt.Fprintf(w, "Lütfen geçerli bir isim giriniz.")
		return
	}

	if pin == "" {
		pin = "1234"
	}

	miktarSayi, _ := strconv.ParseFloat(bakiyeStr, 64)
	baslangicKurus := int(miktarSayi * 100)

	sorgu := "INSERT INTO hesaplar (isim, bakiye, pin) VALUES (?, ?, ?)"
	sonuc, err := config.DB.Exec(sorgu, isim, baslangicKurus, pin)
	if err != nil {
		fmt.Fprintf(w, "Hesap oluşturulurken hata oluştu: %v", err)
		return
	}

	yeniID, _ := sonuc.LastInsertId()
	fmt.Fprintf(w, "Hesap başarıyla oluşturuldu!\nHesap ID: %d\nSahibi: %s\nPIN: %s", yeniID, isim, pin)
}

// ParaYatir: /para-yatir?id=1&miktar=500
func ParaYatir(w http.ResponseWriter, r *http.Request) {
	kullaniciID := r.URL.Query().Get("id")
	miktarStr := r.URL.Query().Get("miktar")

	miktarSayi, err := strconv.ParseFloat(miktarStr, 64)
	if err != nil || miktarSayi <= 0 {
		fmt.Fprintf(w, "Geçerli bir miktar giriniz")
		return
	}

	yatirilacakKurus := int(miktarSayi * 100)

	// Bakiye artırma işlemi
	sonuc, err := config.DB.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", yatirilacakKurus, kullaniciID)
	if err != nil {
		fmt.Fprintf(w, "Para yatırılırken hata oluştu: %v", err)
		return
	}

	etkilenenSatir, _ := sonuc.RowsAffected()
	if etkilenenSatir == 0 {
		fmt.Fprintf(w, "Hesap bulunamadı! ID numarasını kontrol ediniz.")
		return
	}

	// İşlem geçmişine (islemler) ekleme (gonderen_id NULL veya kendisi geçebilir)
	config.DB.Exec("INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)", kullaniciID, kullaniciID, yatirilacakKurus, "YATIRMA")

	fmt.Fprintf(w, "Para yatırma başarılı! %.2f TL yatırıldı.", miktarSayi)
}
