# 🏦 Finansal Backend REST API (Go & SQLite)

Sürdürülebilir, güvenli ve modern backend mimarisi prensipleriyle geliştirilmiş; **Go (Golang)** ve **SQLite** tabanlı bir **Finansal Bankacılık REST API** projesi.

Bu proje; bakiye yönetimi, hesap açma, transactional (ACID) para transferi ve detaylı dekont/işlem geçmişi gibi temel bankacılık operasyonlarını **JSON Body (POST)** ve **Standart REST API Yanıt Formatı** ile sunar.

---

## 🚀 Öne Çıkan Özellikler

* **🔒 Güvenli Veri İletimi:** Hassas kullanıcı verileri (PIN, hesap id, transfer tutarları) URL parametrelerinde taşınmaz; tamamen `POST` metoduna bağlı HTTP Body (JSON) üzerinden işlenir.
* **🛡️ Transactional Transfer (ACID):** Hesaptan hesaba para transferleri SQL Transaction (`tx.Begin()`, `Commit()`, `Rollback()`) mekanizmasıyla yönetilir. Olası bir sistem hatasında bakiye kaybı önlenir.
* **📊 Hassas Bakiye Yönetimi:** Veritabanında bakiyeler `kuruş` (tamsayı/integer) cinsinden saklanarak floating-point (ondalık sayı) hesaplama hataları engellenir. İstemciye (Client) `TL` cinsinden sunulur.
* **📜 Dekont & İşlem Geçmişi (`/gecmis`):** Kullanıcının dahil olduğu gelen/giden tüm transferler ve para yatırma işlemleri zaman damgasıyla listelenir.
* **🎯 Standart JSON Response Formatı:** Tüm yanıtlar HTTP durum kodlarına (`200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Error`) uygun olarak merkezi bir JSON şablonunda dönülür.

---

## 🛠️ Proje Mimarisi (Klasör Yapısı)

Proje, sorumlulukların ayrılması (Separation of Concerns) ilkesine uygun olarak katmanlı bir yapıda kurgulanmıştır:

```text
banka-projesi/
├── config/
│   └── database.go      # SQLite bağlantısı ve tablo ilklendirmeleri (DDL)
├── models/
│   ├── hesap.go         # Hesap modeli ve hesap istek/yanıt struct'ları
│   ├── islem.go         # İşlem (dekont) modeli ve transfer/geçmiş struct'ları
│   └── response.go      # Standart APIResponse modeli
├── handlers/
│   ├── hesap_handler.go    # /bakiye, /hesap-ac, /para-yatir HTTP uç noktaları
│   ├── transfer_handler.go # /para-gonder (Transaction) HTTP uç noktası
│   └── gecmis_handler.go   # /gecmis (İşlem Dökümü) HTTP uç noktası
├── utils/
│   └── response.go      # JSON çıktılarını merkezileştiren yardımcı paket
├── main.go              # Rota tanımlamaları ve HTTP sunucusu başlatma
└── banka.db             # Otomatik oluşturulan SQLite veritabanı
📋 API Endpoints ve Kullanım Kılavuzu
Tüm istekler POST metodu ile ve Content-Type: application/json başlığıyla gönderilmelidir.

1. Bakiye Sorgulama
Endpoint: POST /bakiye

Request Body:

JSON
{
  "id": 1,
  "pin": "1234"
}
Response (200 OK):

JSON
{
  "success": true,
  "message": "Bakiye başarıyla sorgulandı",
  "data": {
    "bakiye": 1500.5,
    "id": 1,
    "isim": "Kaan Çakırer"
  }
}
2. Yeni Hesap Açma
Endpoint: POST /hesap-ac

Request Body:

JSON
{
  "isim": "Ahmet Yılmaz",
  "bakiye": 500.0,
  "pin": "4321"
}
Response (201 Created):

JSON
{
  "success": true,
  "message": "Hesap başarıyla oluşturuldu",
  "data": {
    "id": 2,
    "isim": "Ahmet Yılmaz",
    "bakiye": 500
  }
}
3. Para Yatırma (ATM)
Endpoint: POST /para-yatir

Request Body:

JSON
{
  "id": 1,
  "miktar": 250.0
}
Response (200 OK):

JSON
{
  "success": true,
  "message": "Para yatırma işlemi başarılı",
  "data": {
    "hesap_id": 1,
    "yatirilan_miktar": 250
  }
}
4. Hesaptan Hesaba Para Transferi (EFT/Havale)
Endpoint: POST /para-gonder

Request Body:

JSON
{
  "gonderen_id": 1,
  "alici_id": 2,
  "miktar": 100.0,
  "pin": "1234"
}
Response (200 OK):

JSON
{
  "success": true,
  "message": "Transfer işlemi başarıyla tamamlandı",
  "data": {
    "alici_id": 2,
    "gonderen_id": 1,
    "miktar": 100
  }
}
5. Hesap İşlem Geçmişi (Dekont)
Endpoint: POST /gecmis

Request Body:

JSON
{
  "id": 1,
  "pin": "1234"
}
Response (200 OK):

JSON
{
  "success": true,
  "message": "İşlem geçmişi başarıyla getirildi",
  "data": [
    {
      "id": 3,
      "gonderen_id": 1,
      "alici_id": 2,
      "miktar_tl": 100,
      "islem_tipi": "TRANSFER",
      "tarih": "2026-07-21 14:00:00"
    },
    {
      "id": 2,
      "gonderen_id": 1,
      "alici_id": 1,
      "miktar_tl": 250,
      "islem_tipi": "YATIRMA",
      "tarih": "2026-07-21 13:50:00"
    }
  ]
}
⚡ Çalıştırma Kurulumu
Repoyu klonlayın:

Bash
git clone [https://github.com/kaan-cakirer/banka-projesi.git](https://github.com/kaan-cakirer/banka-projesi.git)
cd banka-projesi
Bağımlılıkları indirin (SQLite sürücüsü vb.):

Bash
go mod tidy
Uygulamayı başlatın:

Bash
go run .
(Sunucu http://localhost:8080 portunda çalışmaya başlayacak ve ilk çalıştırmada banka.db otomatik kurulacaktır.)

🗺️ Gelecek Yol Haritası (Roadmap)
[x] ACID Standartlarında Bakiye Transferi

[x] PIN Kontrol Güvenlik Katmanı

[x] REST API Mimarisi (POST + JSON Body)

[x] İşlem Geçmişi (Dekont) Servisi

[ ] Canlı Döviz Alım/Satım & Portföy Yönetimi (/doviz-al, /doviz-sat)

[ ] Vadeli Mevduat & Faiz Hesabı Servisi

[ ] JWT (JSON Web Token) İle Oturum Yönetimi

[ ] Web Frontend Paneli (HTML / CSS / JS Fetch Entegrasyonu)
