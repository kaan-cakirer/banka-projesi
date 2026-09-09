package usecase

import (
	"errors"
	"math"

	"banka-projesi/models"
	"banka-projesi/repository"
)

// Sentinel error'lar: handler bunları errors.Is ile yakalayıp doğru HTTP
// status kodunu seçecek. Mesaj metnine bakıp karar vermekten kaçınıyoruz.
var (
	ErrGecersizIsim       = errors.New("lütfen geçerli bir isim giriniz")
	ErrNegatifBakiye      = errors.New("başlangıç bakiyesi negatif olamaz")
	ErrHesapBulunamadi    = errors.New("hesap bulunamadı")
	ErrHataliPin          = errors.New("hatalı pin kodu")
	ErrOlusturmaHatasi    = errors.New("hesap oluşturulurken hata oluştu")
	ErrGecersizMiktar     = errors.New("geçersiz miktar")
	ErrBasarisizIslem     = errors.New("işlem başarasız")
	ErrGecmisAlinamadı    = errors.New("geçmiş alınırken hata oluştu")
	ErrYetersizBakiye     = errors.New("yetersiz bakiye")
	ErrGonderenBulunamadı = errors.New("gönderen hesap bulunamadı")
	ErrAliciBulunamadı    = errors.New("alici hesap bulunamadı")
)

// HesapUsecase: iş kurallarının interface'i.
// Handler bu interface'i bilir, ne HTTP ne de SQL'den haberi olur.
type HesapUsecase interface {
	HesapAc(isim string, bakiyeTL float64, pin string) (*models.Hesap, error)
	BakiyeSorgula(id int, pin string) (*models.Hesap, error)
	ParaYatir(id int, miktarTL float64, pin string) (*models.Hesap, error)
	ParaCek(id int, miktarTL float64, pin string) (*models.Hesap, error)
	Transfer(GonderenID int, AliciID int, miktarTL float64, pin string) (*models.TransferIstegi, error)
	IslemSorgula(id int, pin string) ([]models.Islem, error)
}

type hesapUsecase struct {
	repo repository.HesapRepository
}

// NewHesapUsecase: constructor. main.go'da repository'yi buraya vereceğiz.
func NewHesapUsecase(repo repository.HesapRepository) HesapUsecase {
	return &hesapUsecase{repo: repo}
}

// HesapAc: yeni hesap açma iş kuralları.
func (u *hesapUsecase) HesapAc(isim string, bakiyeTL float64, pin string) (*models.Hesap, error) {
	if isim == "" {
		return nil, ErrGecersizIsim
	}

	// Eski handler'daki davranışı koruyoruz: pin boşsa varsayılan atanıyor.
	if pin == "" {
		pin = "1234"
	}

	if bakiyeTL < 0 {
		return nil, ErrNegatifBakiye
	}

	bakiyeKurus := int64(math.Round(bakiyeTL * 100))

	yeniID, err := u.repo.Create(isim, bakiyeKurus, pin)
	if err != nil {
		return nil, ErrOlusturmaHatasi
	}

	return &models.Hesap{
		ID:     yeniID,
		Isim:   isim,
		Bakiye: bakiyeTL,
	}, nil
}

// BakiyeSorgula: PIN doğrulama + bakiye döndürme iş kuralı.
func (u *hesapUsecase) BakiyeSorgula(id int, pin string) (*models.Hesap, error) {
	kayit, err := u.repo.GetByID(id)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	return &models.Hesap{
		ID:     kayit.ID,
		Isim:   kayit.Isim,
		Bakiye: float64(kayit.BakiyeKurus) / 100.0,
	}, nil
}

func (u *hesapUsecase) ParaYatir(id int, miktarTL float64, pin string) (*models.Hesap, error) {

	if miktarTL <= 0 {
		return nil, ErrGecersizMiktar
	}

	kayit, err := u.repo.GetByID(id)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	miktarKurus := int64(math.Round(miktarTL * 100))

	if err := u.repo.BakiyeArttir(id, miktarKurus); err != nil {
		return nil, ErrBasarisizIslem
	}

	return &models.Hesap{
		ID:     kayit.ID,
		Isim:   kayit.Isim,
		Bakiye: float64(kayit.BakiyeKurus)/100 + miktarTL,
	}, nil
}

func (u *hesapUsecase) ParaCek(id int, miktarTL float64, pin string) (*models.Hesap, error) {

	if miktarTL <= 0 {
		return nil, ErrGecersizMiktar
	}

	kayit, err := u.repo.GetByID(id)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	miktarKurus := int64(math.Round(miktarTL * 100))

	if kayit.BakiyeKurus < miktarKurus {
		return nil, ErrYetersizBakiye
	}

	if err := u.repo.BakiyeAzalt(id, miktarKurus); err != nil {
		return nil, ErrBasarisizIslem
	}

	return &models.Hesap{
		ID:     kayit.ID,
		Isim:   kayit.Isim,
		Bakiye: float64(kayit.BakiyeKurus)/100 - miktarTL,
	}, nil
}

func (u *hesapUsecase) Transfer(GonderenID int, AliciID int, miktarTL float64, pin string) (*models.TransferIstegi, error) {

	if miktarTL <= 0 {
		return nil, ErrGecersizMiktar
	}

	if miktarTL <= 0 {
		return nil, ErrGecersizMiktar
	}

	kayitG, err := u.repo.GetByID(GonderenID)
	if err != nil {
		return nil, ErrGonderenBulunamadı
	}

	kayitA, err := u.repo.GetByID(AliciID)
	if err != nil {
		return nil, ErrAliciBulunamadı
	}

	if kayitG.Pin != pin {
		return nil, ErrHataliPin
	}

	miktarKurus := int64(math.Round(miktarTL * 100))

	if kayitG.BakiyeKurus < miktarKurus {
		return nil, ErrYetersizBakiye
	}

	if err := u.repo.Transfer(GonderenID, AliciID, miktarKurus); err != nil {
		return nil, ErrBasarisizIslem
	}

	return &models.TransferIstegi{
		GonderenID: kayitG.ID,
		AliciID:    kayitA.ID,
		Miktar:     miktarTL,
	}, nil

}

func (u *hesapUsecase) IslemSorgula(id int, pin string) ([]models.Islem, error) {

	kayit, err := u.repo.GetByID(id)

	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if pin != kayit.Pin {
		return nil, ErrHataliPin
	}

	islemler, err := u.repo.IslemGecmisi(id)

	if err != nil {
		return nil, ErrGecmisAlinamadı
	}
	return islemler, nil
}
