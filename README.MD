# Banka Projesi (Go + PostgreSQL + Web UI)

Go ile yazılmış, temel bankacılık işlemlerini ve döviz alım/satım akışını sunan bir REST API ve buna bağlı tek sayfa web arayüzü.

Proje şu anda PostgreSQL kullanır, API endpointleri /api altında çalışır ve static klasöründeki frontend doğrudan aynı sunucudan servis edilir.

## Özellikler

- Hesap açma ve PIN ile bakiye doğrulama
- Para yatırma
- Hesaptan hesaba transfer (transaction ile)
- Döviz alım ve döviz satım işlemleri
- Hesaptaki döviz varlıklarını listeleme
- Son 10 işlemi (TL + döviz işlemleri birleşik) listeleme
- Standart JSON response formatı: success, message, data

## Teknoloji

- Go 1.26.5
- PostgreSQL sürücüsü: github.com/lib/pq
- Frontend: HTML + CSS + Vanilla JS (fetch)

## Proje Yapısı

```text
banka-projesi/
├── main.go
├── go.mod
├── config/
│   └── database.go
├── handlers/
│   ├── hesap_handler.go
│   ├── transfer_handler.go
│   ├── gecmis_handler.go
│   ├── doviz_handler.go
│   └── faiz_handler.go
├── models/
│   ├── hesap.go
│   ├── islem.go
│   ├── doviz.go
│   └── response.go
├── services/
│   ├── doviz_service.go
│   └── faiz_service.go
├── utils/
│   ├── response.go
│   └── doviz.go
└── static/
    ├── index.html
    ├── style.css
    └── app.js
```

## Çalışma Mantığı

- Tüm API endpointleri POST kabul eder.
- CORS middleware tüm endpointler için aktif (Access-Control-Allow-Origin: *).
- Kuruş/cents bazlı integer saklama kullanılır, istemciye TL olarak dönülür.
- Para transferi ve döviz işlemleri transaction ile yönetilir.
- Frontend için root path static dosya sunar; /api ile başlayan eşleşmeyen pathler 404 döner.

## Çalıştırma

1. PostgreSQL başlatın.

Örnek Docker komutu:

```bash
docker run --name banka-postgres \
  -e POSTGRES_PASSWORD=gizlisifre \
  -e POSTGRES_DB=bankadb \
  -p 5433:5432 \
  -d postgres:16
```

2. Veritabanı bağlantısını kontrol edin.

config/database.go içinde mevcut bağlantı:

```text
user=postgres password=gizlisifre dbname=bankadb sslmode=disable host=localhost port=5433
```

3. Projeyi çalıştırın.

```bash
go mod tidy
go run .
```

Sunucu: http://localhost:8080

Web arayüz: http://localhost:8080

## API Endpointleri

Tüm endpointler için header:

```text
Content-Type: application/json
```

1. Hesap Aç

Endpoint:

```text
POST /api/hesap-ac
```

Body:

```json
{
  "isim": "Ahmet Yılmaz",
  "bakiye": 500.0,
  "pin": "4321"
}
```

2. Bakiye Sorgula

Endpoint:

```text
POST /api/bakiye
```

Body:

```json
{
  "id": 1,
  "pin": "4321"
}
```

3. Para Yatır

Endpoint:

```text
POST /api/para-yatir
```

Body:

```json
{
  "id": 1,
  "miktar": 250.0,
  "pin": "4321"
}
```

4. Para Gönder

Endpoint:

```text
POST /api/para-gonder
```

Body:

```json
{
  "gonderen_id": 1,
  "alici_id": 2,
  "miktar": 100.0,
  "pin": "4321"
}
```

5. İşlem Geçmişi

Endpoint:

```text
POST /api/gecmis
```

Body:

```json
{
  "id": 1,
  "pin": "4321"
}
```

6. Döviz Al

Endpoint:

```text
POST /api/doviz-al
```

Body:

```json
{
  "hesap_id": 1,
  "doviz_kodu": "USD",
  "miktar": 25.5,
  "pin": "4321"
}
```

7. Döviz Sat

Endpoint:

```text
POST /api/doviz-sat
```

Body:

```json
{
  "hesap_id": 1,
  "doviz_kodu": "USD",
  "miktar": 10.0,
  "pin": "4321"
}
```

8. Döviz Varlıkları

Endpoint:

```text
POST /api/varliklar
```

Body:

```json
{
  "id": 1,
  "pin": "4321"
}
```

## Standart Response Formatı

Başarılı veya hatalı tüm yanıtlar tek formatta döner:

```json
{
  "success": true,
  "message": "Açıklama",
  "data": {}
}
```

## Önemli Notlar

- static/app.js varsayılan olarak API_BASE = /api kullanır. Aynı origin dışında backend kullanacaksanız arayüzde ayarlar (dişli ikon) ile URL değiştirilebilir.
- utils/doviz.go içinde kur bilgisi dış servisten çekilir: https://open.er-api.com/v6/latest/TRY
- handlers/faiz_handler.go ve services altındaki dosyalar şu anda aktif iş mantığı içermeyen placeholder durumundadır.
- Repository içinde banka.db dosyası bulunsa da aktif veritabanı PostgreSQL'dir.

## Hızlı Test (curl)

```bash
curl -X POST http://localhost:8080/api/hesap-ac \
  -H "Content-Type: application/json" \
  -d '{"isim":"Test Kullanici","bakiye":1000,"pin":"1234"}'
```

```bash
curl -X POST http://localhost:8080/api/bakiye \
  -H "Content-Type: application/json" \
  -d '{"id":1,"pin":"1234"}'
```
