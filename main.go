package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"banka-projesi/config"
	"banka-projesi/handlers"
	"banka-projesi/repository"
	"banka-projesi/usecase"
)

// corsMiddleware: Arayüzden (Frontend) gelen isteklere CORS izni sağlar
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	// Veritabanı bağlantısı
	db := config.InitDB()

	hesapRepo := repository.NewHesapRepository(db)
	hesapUsecase := usecase.NewHesapUsecase(hesapRepo)
	hesapHandler := handlers.NewHesapHandler(hesapUsecase)

	dovizRepo := repository.NewDovizRepository(db)
	dovizUsecase := usecase.NewDovizUsecase(dovizRepo, hesapRepo)
	dovizHandler := handlers.NewDovizHandler(dovizUsecase)

	// 1. API Endpoints
	http.HandleFunc("/api/hesap-ac", corsMiddleware(hesapHandler.HesapAc))
	http.HandleFunc("/api/bakiye", corsMiddleware(hesapHandler.BakiyeSorgula))
	http.HandleFunc("/api/para-yatir", corsMiddleware(hesapHandler.ParaYatir))
	http.HandleFunc("/api/para-cek", corsMiddleware(hesapHandler.ParaCek))
	http.HandleFunc("/api/islem-gecmisi", corsMiddleware(hesapHandler.IslemGecmisi))
	http.HandleFunc("/api/para-gonder", corsMiddleware(hesapHandler.ParaGonder))
	http.HandleFunc("/api/doviz-al", corsMiddleware(dovizHandler.DovizAl))
	http.HandleFunc("/api/doviz-sat", corsMiddleware(dovizHandler.DovizSat))
	http.HandleFunc("/api/doviz-kur", corsMiddleware(dovizHandler.DovizKur))
	http.HandleFunc("/api/varliklar", corsMiddleware(dovizHandler.VarliklariGetir))

	// 2. Frontend Statik Dosya Sunucusu (Çakışmayı önleyen özel yönlendirme)
	fs := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// İstek /api/ ile başlıyorsa ve yukarıdaki handler'lara takılmadıysa 404 JSON/Not Found bas
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Tarayıcı Last-Modified'a bakıp sezgisel önbellek uyguluyor ve güncellenen
		// CSS/JS'i almıyordu; no-cache ile her istekte doğrulamaya zorluyoruz.
		w.Header().Set("Cache-Control", "no-cache")

		// Aksi halde static klasöründeki dosyaları (index.html, style.css, app.js) sun
		fs.ServeHTTP(w, r)
	})

	fmt.Println("🚀 Banka REST API ve Web Arayüzü http://localhost:8080 adresinde çalışıyor...")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Sunucu başlatılamadı: ", err)
	}
}
