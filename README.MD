# 🏦 Banka Projesi (Go REST API)

Go (Golang) dili kullanılarak geliştirilmiş, Clean Architecture (Katmanlı Mimari) prensiplerine uygun, SQLite veritabanı destekli güvenli bir bankacılık backend sistemidir.

Proje, ACID uyumlu veritabanı işlemleri (Transactions), güvenli bakiye transferleri, dinamik hesap oluşturma ve işlem kayıt loglaması sunar.

---

## 🏗️ Proje Mimarısı

Proje, sorumlulukların ayrılması (Separation of Concerns) ilkesine dayanarak katmanlara bölünmüştür:

```text
banka-projesi/
├── config/        # Veritabanı (SQLite) bağlantısı ve tablo ilklendirmeleri
├── handlers/      # HTTP İsteklerini karşılayan uç noktalar (Endpoints)
├── models/        # Veri yapıları ve Struct tanımları
├── services/      # İş mantığı ve dış servis entegrasyonları (Döviz, Faiz vb.)
├── main.go        # Uygulama giriş noktası ve HTTP rotaları (Routes)
├── go.mod         # Proje bağımlılıkları
└── banka.db       # SQLite Veritabanı dosyası