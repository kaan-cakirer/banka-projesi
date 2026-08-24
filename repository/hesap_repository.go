package repository

import (
	"banka-projesi/models"
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
	BakiyeArttir(id int, miktarKurus int64) error
	IslemGecmisi(id int) ([]models.Islem, error)
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

func (r *postgresHesapRepository) BakiyeArttir(id int, miktarKurus int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `UPDATE hesaplar SET bakiye = bakiye + $1 WHERE id = $2`
	if _, err = tx.Exec(query, miktarKurus, id); err != nil {
		tx.Rollback()
		return err
	}

	query = `INSERT INTO islemler (gonderen_id, alici_id, miktar_tl, islem_tipi) VALUES ($1, $2, $3, $4)`
	if _, err = tx.Exec(query, id, id, miktarKurus, "YATIRMA"); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *postgresHesapRepository) IslemGecmisi(id int) ([]models.Islem, error) {
	sorgu := `
	SELECT id, gonderen_id, alici_id, miktar_tl as miktar, islem_tipi, tarih, NULL as doviz_kodu, NULL as doviz_miktar
	FROM islemler
	WHERE gonderen_id = $1 OR alici_id = $2
	UNION ALL
	SELECT id, hesap_id as gonderen_id, NULL as alici_id, harcanan_tl_kurus as miktar, islem_tipi, tarih, doviz_kodu, miktar_cent as doviz_miktar
	FROM doviz_islemleri
	WHERE hesap_id = $3
	ORDER BY tarih DESC
	LIMIT 10`
	satirlar, err := r.db.Query(sorgu, id, id, id)
	if err != nil {
		return nil, err
	}
	defer satirlar.Close()

	var islemler []models.Islem

	for satirlar.Next() {
		var (
			islemID     int
			gonderenID  int
			aliciID     sql.NullInt64
			miktarKurus int64
			islemTipi   string
			tarih       string
			dovizKodu   sql.NullString
			dovizMiktar sql.NullFloat64
		)
		if err := satirlar.Scan(&islemID, &gonderenID, &aliciID, &miktarKurus, &islemTipi, &tarih, &dovizKodu, &dovizMiktar); err != nil {
			return nil, err
		}
		islem := models.Islem{
			ID:         islemID,
			GonderenID: gonderenID,
			MiktarTL:   float64(miktarKurus) / 100.0,
			IslemTipi:  islemTipi,
			Tarih:      tarih,
		}

		if aliciID.Valid {
			islem.AliciID = int(aliciID.Int64)
		}
		if dovizKodu.Valid {
			islem.DovizKodu = dovizKodu.String
		}
		if dovizMiktar.Valid {
			islem.DovizMiktar = dovizMiktar.Float64 / 100.0
		}
		islemler = append(islemler, islem)
	}
	if err := satirlar.Err(); err != nil {
		return nil, err
	}
	return islemler, nil

}
