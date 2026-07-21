package main

import (
	"fmt"
	"log"
	"net/http"

	// Kendi paketlerimizi içe aktarıyoruz
	"banka-projesi/config"
	"banka-projesi/handlers"
)

func main() {
	db := config.InitDB()

	defer db.Close()

	http.HandleFunc("/hesap-ac", handlers.HesapAc)
	http.HandleFunc("/bakiye", handlers.BakiyeSorgula)
	http.HandleFunc("/para-yatir", handlers.ParaYatir)
	http.HandleFunc("/para-gonder", handlers.ParaGonder)
	http.HandleFunc("/islem-gecmisi", handlers.IslemGecmisi)

	fmt.Println("Banka sunucusu 8080 portunda dinliyor...")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Sunucu başlatılamadı: ", err)
	}

}
