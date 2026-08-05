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

	varliklarTablosu := `
    CREATE TABLE IF NOT EXISTS varliklar (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        hesap_id INTEGER NOT NULL,
        doviz_kodu TEXT NOT NULL, 
        miktar INTEGER NOT NULL DEFAULT 0,
        FOREIGN KEY(hesap_id) REFERENCES hesaplar(id),
        UNIQUE(hesap_id, doviz_kodu)
    );`

	dovizIslemleriTablosu := `
    CREATE TABLE IF NOT EXISTS doviz_islemleri (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        hesap_id INTEGER NOT NULL,
        doviz_kodu TEXT NOT NULL,
        miktar_cent INTEGER NOT NULL, 
        harcanan_tl_kurus INTEGER NOT NULL, 
        kur_fiyati REAL NOT NULL,          
        islem_tipi TEXT NOT NULL,          
        tarih DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(hesap_id) REFERENCES hesaplar(id)
    );`

	_, err = DB.Exec(hesaplarTablosu)
	if err != nil {
		log.Fatal("Hesaplar tablosu oluşturulurken hata: ", err)
	}

	_, err = DB.Exec(islemlerTablosu)
	if err != nil {
		log.Fatal("İşlemler tablosu oluşturulurken hata: ", err)
	}

	_, err = DB.Exec(varliklarTablosu)
	if err != nil {
		log.Fatal("Varlıklar tablosu oluşturulurken hata: ", err)
	}

	_, err = DB.Exec(dovizIslemleriTablosu)
	if err != nil {
		log.Fatal("Döviz İşlemleri Tablosu Oluşturulurken Hata: ", err)
	}

	fmt.Println("Veritabanı Bağlantısı Başarıyla Kuruldu")
	return DB
}
