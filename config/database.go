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

	tabloSorgusu := `
	CREATE TABLE IF NOT EXISTS hesaplar (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		isim TEXT NOT NULL,
		bakiye INTEGER DEFAULT 0
	);`

	_, err = DB.Exec(tabloSorgusu)
	if err != nil {
		log.Fatal("Tablo Oluşturulurken Hata Oluştu: ", err)
	}

	fmt.Println("Veritabanı Bağlantısı Başarıyla Kuruldu")
	return DB
}
