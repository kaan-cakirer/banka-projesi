package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() *sql.DB {
	var err error

	DB, err = sql.Open("sqlite", "banka.db")
	if err != nil {
		log.Fatal("Veritabanı Açılmadı: ", err)

	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("Veritabanına Ulaşılamıyor: ", err)
	}

	hesaplarTablosu := `
	CREATE TABLE IF NOT EXISTS hesaplar (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		isim TEXT NOT NULL,
		bakiye INTEGER DEFAULT 0,
		pin TEXT DEFAULT '1234'
	);`

	islemlerTablosu := `
	CREATE TABLE IF NOT EXISTS islemler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		gonderen_id INTEGER,
		alici_id INTEGER,
		miktar INTEGER NOT NULL,
		islem_tipi TEXT NOT NULL,
		tarih DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(hesaplarTablosu)
	if err != nil {
		log.Fatal("Hesaplar tablosu oluşturulurken hata: ", err)
	}

	_, err = DB.Exec(islemlerTablosu)
	if err != nil {
		log.Fatal("İşlemler tablosu oluşturulurken hata: ", err)
	}

	fmt.Println("Veritabanı Bağlantısı Başarıyla Kuruldu")
	return DB
}
