# Banka Projesi

Go ve PostgreSQL ile geliştirilmiş, temel bankacılık işlemlerini ve döviz alım/satım akışını sunan bir REST API. `static/` klasöründeki tek sayfa web arayüzü de aynı Go sunucusundan servis edilir.

## Özellikler

- Hesap açma ve 4 haneli PIN ile giriş/bakiye doğrulama
- Para yatırma
- Hesaptan hesaba transfer (transaction ile)
- Döviz alım ve döviz satım işlemleri
- Hesaptaki döviz varlıklarını listeleme
- İşlem geçmişini (TL + döviz işlemleri) listeleme
- Standart JSON response formatı: success, message, data

## Teknoloji

- Go 1.26.5 veya üzeri
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

- Tüm API endpointleri `POST` kabul eder; istek ve yanıt gövdeleri JSON'dur.
- CORS middleware tüm endpointler için aktiftir (`Access-Control-Allow-Origin: *`).
- TL ve döviz miktarları veritabanında kuruş/cent bazlı integer olarak saklanır.
- Para transferi ve döviz işlemleri PostgreSQL transaction'ı içinde yürütülür.
- `/` yolu `static/` klasöründeki frontend'i sunar; eşleşmeyen `/api/` yolları 404 döner.
- Döviz kurları `https://open.er-api.com/v6/latest/TRY` servisinden alınır. Döviz işlemleri için internet bağlantısı gerekir.

## Çalıştırma

### Gereksinimler

- Go 1.26.5 veya üzeri
- PostgreSQL 16 veya üzeri
- Döviz işlemleri kullanılacaksa internet bağlantısı

### PostgreSQL'i başlatma

Örnek Docker komutu:

```bash
docker run --name banka-postgres \
  -e POSTGRES_PASSWORD=gizlisifre \
  -e POSTGRES_DB=bankadb \
  -p 5433:5432 \
  -d postgres:16
```

Bu Docker komutu veritabanını `localhost:5433` üzerinde açar. Uygulamanın bağlantı bilgileri [config/database.go](config/database.go) içinde sabit olarak tanımlıdır:

```text
user=postgres password=gizlisifre dbname=bankadb sslmode=disable host=localhost port=5433
```

Farklı bağlantı bilgileri kullanacaksanız `config/database.go` içindeki connection string'i güncelleyin.

### Uygulamayı çalıştırma

```bash
go mod tidy
go run .
```

Sunucu ve web arayüzü: http://localhost:8080

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
POST /api/islem-gecmisi
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

- `static/app.js` varsayılan olarak `API_BASE = /api` kullanır. Backend'i farklı bir adreste çalıştıracaksanız arayüzdeki ayarlar üzerinden API adresini değiştirebilirsiniz.
- PIN'ler frontend'de kalıcı olarak saklanmaz; sayfa yenilendiğinde oturum kapanır.
- Uygulama başlarken gerekli PostgreSQL tablolarını `CREATE TABLE IF NOT EXISTS` ile oluşturur.
- [handlers/faiz_handler.go](handlers/faiz_handler.go) ve [services/faiz_service.go](services/faiz_service.go) mevcut olsa da faiz akışı `main.go` içinde route edilmemiştir.
- Eski bir `banka.db` dosyası bulunsa bile aktif veritabanı PostgreSQL'dir.

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

## Lisans

Bu proje için henüz bir lisans belirtilmemiştir.
