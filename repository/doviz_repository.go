package repository

import "database/sql"

// VarlikKaydi: repository'nin dışarıya verdiği ham veri.
// DB'deki gerçek haliyle birebir - miktar burada CENT (int64).
type VarlikKaydi struct {
	ID         int
	HesapID    int
	DovizKodu  string
	MiktarCent int64
}

// DovizRepository: döviz alım/satım/varlık veritabanı işlemlerinin interface'i.
type DovizRepository interface {
	VarlikGetir(hesapID int, dovizKodu string) (*VarlikKaydi, error)
	VarliklarGetir(hesapID int) ([]VarlikKaydi, error)
	DovizAl(hesapID int, dovizKodu string, dovizCent int64, harcananKurus int64, kurFiyati float64) error
	DovizSat(hesapID int, dovizKodu string, dovizCent int64, kazanilanKurus int64, kurFiyati float64) error
}

type postgresDovizRepository struct {
	db *sql.DB
}

// NewDovizRepository: constructor. main.go'da config.DB'yi buraya vereceğiz.
func NewDovizRepository(db *sql.DB) DovizRepository {
	return &postgresDovizRepository{db: db}
}

func (r *postgresDovizRepository) VarlikGetir(hesapID int, dovizKodu string) (*VarlikKaydi, error) {
	query := `SELECT id, hesap_id, doviz_kodu, miktar FROM varliklar WHERE hesap_id = $1 AND doviz_kodu = $2`

	kayit := &VarlikKaydi{}
	err := r.db.QueryRow(query, hesapID, dovizKodu).Scan(&kayit.ID, &kayit.HesapID, &kayit.DovizKodu, &kayit.MiktarCent)
	if err != nil {
		return nil, err
	}
	return kayit, nil
}

func (r *postgresDovizRepository) VarliklarGetir(hesapID int) ([]VarlikKaydi, error) {
	query := `SELECT id, hesap_id, doviz_kodu, miktar FROM varliklar WHERE hesap_id = $1`

	satirlar, err := r.db.Query(query, hesapID)
	if err != nil {
		return nil, err
	}
	defer satirlar.Close()

	varliklar := make([]VarlikKaydi, 0)
	for satirlar.Next() {
		var v VarlikKaydi
		if err := satirlar.Scan(&v.ID, &v.HesapID, &v.DovizKodu, &v.MiktarCent); err != nil {
			return nil, err
		}
		varliklar = append(varliklar, v)
	}
	if err := satirlar.Err(); err != nil {
		return nil, err
	}
	return varliklar, nil
}

// DovizAl: TL bakiyesini düşer, döviz varlığını artırır, dekont kaydı oluşturur. Tek transaction.
func (r *postgresDovizRepository) DovizAl(hesapID int, dovizKodu string, dovizCent int64, harcananKurus int64, kurFiyati float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye - $1 WHERE id = $2", harcananKurus, hesapID); err != nil {
		tx.Rollback()
		return err
	}

	varlikSorgusu := `
		INSERT INTO varliklar (hesap_id, doviz_kodu, miktar)
		VALUES ($1, $2, $3)
		ON CONFLICT(hesap_id, doviz_kodu)
		DO UPDATE SET miktar = miktar + EXCLUDED.miktar`
	if _, err = tx.Exec(varlikSorgusu, hesapID, dovizKodu, dovizCent); err != nil {
		tx.Rollback()
		return err
	}

	dekontSorgusu := `
		INSERT INTO doviz_islemleri (hesap_id, doviz_kodu, miktar_cent, harcanan_tl_kurus, kur_fiyati, islem_tipi)
		VALUES ($1, $2, $3, $4, $5, 'ALIM')`
	if _, err = tx.Exec(dekontSorgusu, hesapID, dovizKodu, dovizCent, harcananKurus, kurFiyati); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// DovizSat: TL bakiyesini artırır, döviz varlığını düşer (sıfırsa siler), dekont kaydı oluşturur. Tek transaction.
func (r *postgresDovizRepository) DovizSat(hesapID int, dovizKodu string, dovizCent int64, kazanilanKurus int64, kurFiyati float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec("UPDATE hesaplar SET bakiye = bakiye + $1 WHERE id = $2", kazanilanKurus, hesapID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("UPDATE varliklar SET miktar = miktar - $1 WHERE hesap_id = $2 AND doviz_kodu = $3", dovizCent, hesapID, dovizKodu); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("DELETE FROM varliklar WHERE hesap_id = $1 AND doviz_kodu = $2 AND miktar = 0", hesapID, dovizKodu); err != nil {
		tx.Rollback()
		return err
	}

	dekontSorgusu := `
		INSERT INTO doviz_islemleri (hesap_id, doviz_kodu, miktar_cent, harcanan_tl_kurus, kur_fiyati, islem_tipi)
		VALUES ($1, $2, $3, $4, $5, 'SATIM')`
	if _, err = tx.Exec(dekontSorgusu, hesapID, dovizKodu, dovizCent, kazanilanKurus, kurFiyati); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
