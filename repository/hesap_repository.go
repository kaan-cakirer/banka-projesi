package repository

import (
	"database/sql"
	"errors"
)

// HesapKaydi: repository'nin dışarıya verdiği ham veri.
// DB'deki gerçek haliyle birebir - bakiye burada KURUŞ (int64).
// TL'ye çevirme işi usecase'in görevi, repository bunu bilmez.
type HesapKaydi struct {
	ID          int
	Isim        string
	BakiyeKurus int64
	Pin         string
}

// HesapRepository: veritabanı işlemlerinin interface'i.
// Usecase katmanı bu interface'i bilir, PostgreSQL'i değil.
type HesapRepository interface {
	Create(isim string, bakiyeKurus int64, pin string) (int, error)
	GetByID(id int) (*HesapKaydi, error)
}

type postgresHesapRepository struct {
	db *sql.DB
}

// NewHesapRepository: constructor. main.go'da config.DB'yi buraya vereceğiz.
func NewHesapRepository(db *sql.DB) HesapRepository {
	return &postgresHesapRepository{db: db}
}

// Create: yeni hesap satırı ekler, oluşan ID'yi RETURNING ile geri okur.
// NOT: lib/pq, sonuc.LastInsertId()'i DESTEKLEMEZ - bu yüzden RETURNING id kullanıyoruz.
func (r *postgresHesapRepository) Create(isim string, bakiyeKurus int64, pin string) (int, error) {
	query := `
		INSERT INTO hesaplar (isim, bakiye, pin)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var yeniID int
	err := r.db.QueryRow(query, isim, bakiyeKurus, pin).Scan(&yeniID)
	if err != nil {
		return 0, err
	}
	return yeniID, nil
}

// GetByID: tek hesap satırı okur.
func (r *postgresHesapRepository) GetByID(id int) (*HesapKaydi, error) {
	query := `SELECT id, isim, bakiye, pin FROM hesaplar WHERE id = $1`

	kayit := &HesapKaydi{}
	err := r.db.QueryRow(query, id).Scan(&kayit.ID, &kayit.Isim, &kayit.BakiyeKurus, &kayit.Pin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("hesap bulunamadı")
		}
		return nil, err
	}
	return kayit, nil
}
