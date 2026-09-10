package usecase

import (
	"database/sql"
	"errors"
	"math"

	"banka-projesi/models"
	"banka-projesi/repository"
	"banka-projesi/utils"
)

// Döviz işlemlerine özel sentinel error'lar. Ortak olanlar (ErrHesapBulunamadi,
// ErrHataliPin, ErrGecersizMiktar, ErrYetersizBakiye, ErrBasarisizIslem)
// hesap_usecase.go'da tanımlı, aynı paket olduğu için burada da kullanılıyor.
var (
	ErrGecersizDovizKodu  = errors.New("geçersiz döviz kodu veya kur sunucusuna ulaşılamadı")
	ErrYetersizDoviz      = errors.New("yetersiz döviz bakiyesi")
	ErrVarliklarAlinamadı = errors.New("varlıklar alınırken hata oluştu")
)

// DovizUsecase: döviz alım/satım/varlık sorgulama iş kurallarının interface'i.
type DovizUsecase interface {
	DovizAl(hesapID int, dovizKodu string, miktar float64, pin string) (*models.DovizIslem, error)
	DovizSat(hesapID int, dovizKodu string, miktar float64, pin string) (*models.DovizIslem, error)
	VarliklarGetir(hesapID int, pin string) ([]models.Varlik, error)
	KurGetir(dovizKodu string) (*models.DovizKurYaniti, error)
}

type dovizUsecase struct {
	dovizRepo repository.DovizRepository
	hesapRepo repository.HesapRepository
}

// NewDovizUsecase: constructor. main.go'da repository'leri buraya vereceğiz.
func NewDovizUsecase(dovizRepo repository.DovizRepository, hesapRepo repository.HesapRepository) DovizUsecase {
	return &dovizUsecase{dovizRepo: dovizRepo, hesapRepo: hesapRepo}
}

// DovizAl: PIN doğrulama + TL bakiye kontrolü + döviz alım iş kuralı.
func (u *dovizUsecase) DovizAl(hesapID int, dovizKodu string, miktar float64, pin string) (*models.DovizIslem, error) {
	if miktar <= 0 {
		return nil, ErrGecersizMiktar
	}

	kurOran, err := utils.KurGetir(dovizKodu)
	if err != nil {
		return nil, ErrGecersizDovizKodu
	}

	kayit, err := u.hesapRepo.GetByID(hesapID)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	dovizCent := int64(math.Round(miktar * 100))
	harcananKurus := int64(math.Round(miktar * kurOran * 100))

	if kayit.BakiyeKurus < harcananKurus {
		return nil, ErrYetersizBakiye
	}

	if err := u.dovizRepo.DovizAl(hesapID, dovizKodu, dovizCent, harcananKurus, kurOran); err != nil {
		return nil, ErrBasarisizIslem
	}

	return &models.DovizIslem{
		HesapID:    hesapID,
		DovizKodu:  dovizKodu,
		Miktar:     miktar,
		HarcananTL: float64(harcananKurus) / 100.0,
		KurFiyati:  kurOran,
		IslemTipi:  "ALIM",
	}, nil
}

// DovizSat: PIN doğrulama + döviz bakiye kontrolü + döviz satım iş kuralı.
func (u *dovizUsecase) DovizSat(hesapID int, dovizKodu string, miktar float64, pin string) (*models.DovizIslem, error) {
	if miktar <= 0 {
		return nil, ErrGecersizMiktar
	}

	kurOran, err := utils.KurGetir(dovizKodu)
	if err != nil {
		return nil, ErrGecersizDovizKodu
	}

	kayit, err := u.hesapRepo.GetByID(hesapID)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	dovizCent := int64(math.Round(miktar * 100))

	varlik, err := u.dovizRepo.VarlikGetir(hesapID, dovizKodu)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrYetersizDoviz
		}
		return nil, ErrBasarisizIslem
	}

	if varlik.MiktarCent < dovizCent {
		return nil, ErrYetersizDoviz
	}

	kazanilanKurus := int64(math.Round(miktar * kurOran * 100))

	if err := u.dovizRepo.DovizSat(hesapID, dovizKodu, dovizCent, kazanilanKurus, kurOran); err != nil {
		return nil, ErrBasarisizIslem
	}

	return &models.DovizIslem{
		HesapID:    hesapID,
		DovizKodu:  dovizKodu,
		Miktar:     miktar,
		HarcananTL: float64(kazanilanKurus) / 100.0,
		KurFiyati:  kurOran,
		IslemTipi:  "SATIM",
	}, nil
}

// KurGetir: seçili dövizin güncel TL karşılığını döner, hesaba/PIN'e dokunmaz,
// işlem yapmaz - sadece arayüzde fiyat göstermek için kullanılır.
func (u *dovizUsecase) KurGetir(dovizKodu string) (*models.DovizKurYaniti, error) {
	if dovizKodu == "" {
		return nil, ErrGecersizDovizKodu
	}

	kurOran, err := utils.KurGetir(dovizKodu)
	if err != nil {
		return nil, ErrGecersizDovizKodu
	}

	return &models.DovizKurYaniti{DovizKodu: dovizKodu, BirimTL: kurOran}, nil
}

// VarliklarGetir: PIN doğrulama + hesabın döviz varlıklarını listeleme iş kuralı.
func (u *dovizUsecase) VarliklarGetir(hesapID int, pin string) ([]models.Varlik, error) {
	kayit, err := u.hesapRepo.GetByID(hesapID)
	if err != nil {
		return nil, ErrHesapBulunamadi
	}

	if kayit.Pin != pin {
		return nil, ErrHataliPin
	}

	kayitlar, err := u.dovizRepo.VarliklarGetir(hesapID)
	if err != nil {
		return nil, ErrVarliklarAlinamadı
	}

	varliklar := make([]models.Varlik, 0, len(kayitlar))
	for _, k := range kayitlar {
		varliklar = append(varliklar, models.Varlik{
			ID:        k.ID,
			HesapID:   k.HesapID,
			DovizKodu: k.DovizKodu,
			Miktar:    float64(k.MiktarCent) / 100.0,
		})
	}
	return varliklar, nil
}
