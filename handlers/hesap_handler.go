package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"banka-projesi/config"
)

func BakiyeSorgula(w http.ResponseWriter, r *http.Request) {
	gelenID := r.URL.Query().Get("id")

	var isim string
	var bakiyeKurus int

	sorgu := "SELECT isim, bakiye FROM hesaplar WHERE id = ?"

	err := config.DB.QueryRow(sorgu, gelenID).Scan(&isim, &bakiyeKurus)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(w, "Hesap Bulunamadı")
			return
		}
		fmt.Fprintf(w, "Veritabanı Hatası: %v", err)
	}
	bakiyeTL := float64(bakiyeKurus) / 100.0

	fmt.Fprintf(w, "Hesap Sahibi: %s\nBakiye: %2.f", isim, bakiyeTL)
}

func ParaYatir(w http.ResponseWriter, r *http.Request) {

	kullaniciID := r.URL.Query().Get("id")
	miktarStr := r.URL.Query().Get("miktar")

	miktarSayi, err := strconv.ParseFloat(miktarStr, 64)

	if err != nil || miktarSayi <= 0 {
		fmt.Fprintf(w, "Geçerli bir miktar giriniz")
		return
	}

	yatirilacakKurus := int(miktarSayi * 100)

	sonuc, err := config.DB.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", yatirilacakKurus, kullaniciID)
	if err != nil {
		fmt.Fprintf(w, "Para yatırılırken hata oluştu: %v", err)
		return
	}

	etkilenenSatir, _ := sonuc.RowsAffected()
	if etkilenenSatir == 0 {
		fmt.Fprintf(w, "Hesap bulunamadi! ID numarasini kontrol ediniz.")
		return
	}

	fmt.Fprintf(w, "Para yatırma başarılı! %.2f TL yatırıldı.", miktarSayi)
}

func HesapAc(w http.ResponseWriter, r *http.Request) {
	isim := r.URL.Query().Get("isim")
	bakiyeStr := r.URL.Query().Get("bakiye")

	if isim == "" {
		fmt.Fprintf(w, "Lütfen geçerli bir isim giriniz.")
		return
	}

	miktarSayi, _ := strconv.ParseFloat(bakiyeStr, 64)
	baslangicKurus := int(miktarSayi * 100)

	sorgu := "INSERT INTO HESAPLAR (isim, bakiye) VALUES (?, ?)"

	sonuc, err := config.DB.Exec(sorgu, isim, baslangicKurus)
	if err != nil {
		fmt.Fprintf(w, "Hesap oluşturulurken hata oluştu: %v", err)
		return
	}
	yeniID, _ := sonuc.LastInsertId()
	fmt.Fprintf(w, "Hesap başarıyla oluşturuldu! Hesap ID: %d, Sahibi: %s", yeniID, isim)

}
