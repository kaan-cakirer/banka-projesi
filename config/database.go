package config

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL sürücüsü
)

var DB *sql.DB

func InitDB() *sql.DB {
	// PostgreSQL bağlantı adresi (Docker'da verdiğimiz bilgiler)
	connStr := "user=postgres password=gizlisifre dbname=bankadb sslmode=disable host=localhost port=5433"

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("PostgreSQL'e bağlanılamadı: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("PostgreSQL veritabanına erişilemiyor (Docker çalışıyor mu?): ", err)
	}

	// 1. Hesaplar Tablosu (AUTOINCREMENT yerine SERIAL kullanıyoruz)
	hesaplarTablosu := `
	CREATE TABLE IF NOT EXISTS hesaplar (
		id SERIAL PRIMARY KEY,
		isim TEXT NOT NULL,
		bakiye INTEGER NOT NULL DEFAULT 0,
		pin TEXT NOT NULL
	);`

	// 2. Transfer Geçmişi Tablosu
	islemlerTablosu := `
	CREATE TABLE IF NOT EXISTS islemler (
		id SERIAL PRIMARY KEY,
		gonderen_id INTEGER,
		alici_id INTEGER,
		miktar_tl INTEGER NOT NULL,
		islem_tipi TEXT NOT NULL,
		tarih TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 3. Döviz Varlıkları Tablosu
	varliklarTablosu := `
	CREATE TABLE IF NOT EXISTS varliklar (
		id SERIAL PRIMARY KEY,
		hesap_id INTEGER NOT NULL REFERENCES hesaplar(id),
		doviz_kodu TEXT NOT NULL, 
		miktar INTEGER NOT NULL DEFAULT 0,
		UNIQUE(hesap_id, doviz_kodu)
	);`

	// 4. Döviz İşlem Geçmişi Tablosu
	dovizIslemleriTablosu := `
	CREATE TABLE IF NOT EXISTS doviz_islemleri (
		id SERIAL PRIMARY KEY,
		hesap_id INTEGER NOT NULL REFERENCES hesaplar(id),
		doviz_kodu TEXT NOT NULL,
		miktar_cent INTEGER NOT NULL,
		harcanan_tl_kurus INTEGER NOT NULL,
		kur_fiyati REAL NOT NULL,
		islem_tipi TEXT NOT NULL,
		tarih TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Tabloları çalıştır
	sorgular := []string{hesaplarTablosu, islemlerTablosu, varliklarTablosu, dovizIslemleriTablosu}
	for _, sorgu := range sorgular {
		_, err := DB.Exec(sorgu)
		if err != nil {
			log.Fatalf("Tablo oluşturulamadı: %v\nSorgu: %s", err, sorgu)
		}
	}

	log.Println("🐘 PostgreSQL bağlantısı başarılı ve tablolar hazır!")
	return DB
}
