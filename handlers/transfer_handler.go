package handlers

import (
	"banka-projesi/config"
	"banka-projesi/models"
	"encoding/json"
	"fmt"
	"net/http"
)

// ParaGonder: POST /para-gonder (JSON Body: {"gonderen_id": 1, "alici_id": 2, "miktar": 100, "pin": "1234"})
func ParaGonder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var istek models.TransferIstegi
	err := json.NewDecoder(r.Body).Decode(&istek)
	if err != nil || istek.Miktar <= 0 {
		fmt.Fprintf(w, "Geçersiz JSON verisi veya miktar")
		return
	}

	gonderilecekKurus := int(istek.Miktar * 100)

	var gonderenBakiye int
	var gercekPin string

	sorgu := "SELECT bakiye, pin FROM hesaplar WHERE id = ?"
	err = config.DB.QueryRow(sorgu, istek.GonderenID).Scan(&gonderenBakiye, &gercekPin)

	if err != nil {
		fmt.Fprintf(w, "Gönderen hesap bulunamadı")
		return
	}

	// PIN Kontrolü
	if gercekPin != istek.Pin {
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

	// 1. Gönderenin bakiyesini düş
	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye - ? WHERE id = ?", gonderilecekKurus, istek.GonderenID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(w, "Gönderilen bakiye düşülürken hata oluştu: %v", err)
		return
	}

	// 2. Alıcının bakiyesini artır
	_, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye + ? WHERE id = ?", gonderilecekKurus, istek.AliciID)
	if err != nil {
		tx.Rollback()
		fmt.Fprintf(w, "Gönderilen bakiye yüklenirken hata oluştu: %v", err)
		return
	}

	// 3. İşlem geçmişine kaydet
	islemSorgusu := "INSERT INTO islemler (gonderen_id, alici_id, miktar, islem_tipi) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(islemSorgusu, istek.GonderenID, istek.AliciID, gonderilecekKurus, "TRANSFER")
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

	fmt.Fprintf(w, "Transfer Başarılı!\n%.2f TL tutarındaki miktar %d kullanıcısından %d kullanıcısına başarıyla aktarıldı.", istek.Miktar, istek.GonderenID, istek.AliciID)
}
