package withdraw

import (
	"fmt"
	"time"

	golnurl "github.com/fiatjaf/go-lnurl"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
)

const payWithdrawEndpoint = "https://%s/lnurl/withdraw/pay"
const withdrawRequestEndpoint = "https://%s/lnurl/withdraw/request?k1=%s"

type Manager interface {
	CreateLNURLW(username, walletId, description string, minMsat, maxMsat uint64, maxReuse uint, expiry time.Time) (lnurl string, err error)
	GetLNURLW(username, walletId, k1 string) (lnurl string, err error)
	GetWithdrawRequest(k1 string) (withdraw *models.Withdraw, callback string, err error)
}

type manager struct {
	hostname string
	db       *db.DB
}

func NewManager(hostname string, db *db.DB) Manager {
	return &manager{
		hostname: hostname,
		db:       db,
	}
}

func (m *manager) CreateLNURLW(username, walletId, description string, minMsat, maxMsat uint64, maxReuse uint, expiry time.Time) (lnurl string, err error) {
	if wallet, err := m.db.Repo.GetWallet(m.db.DB, username, walletId); err != nil {
		return "", err
	} else if wallet.Locked {
		return "", models.ErrLockedWallet
	}

	withdraw, err := m.db.Repo.CreateWithdraw(m.db.DB, &models.Withdraw{
		WalletID:    walletId,
		Username:    username,
		Description: description,
		MinMsat:     minMsat,
		MaxMsat:     maxMsat,
		MaxUse:      maxReuse,
		Expiry:      expiry,
	})
	if err != nil {
		return "", err
	} else {
		return m.createInitWithdrawLink(withdraw.K1)
	}
}

func (m *manager) GetLNURLW(username, walletId, k1 string) (lnurl string, err error) {
	withdraw, err := m.db.Repo.GetWalletWithdraw(m.db.DB, username, walletId, k1, false)
	if err != nil {
		return "", err
	} else {
		return m.createInitWithdrawLink(withdraw.K1)
	}
}

func (m *manager) GetWithdrawRequest(k1 string) (withdraw *models.Withdraw, callback string, err error) {
	withdraw, err = m.db.Repo.GetWithdraw(m.db.DB, k1, false)
	if err != nil {
		return nil, "", err
	} else {
		return withdraw, fmt.Sprintf(payWithdrawEndpoint, m.hostname), nil
	}
}

func (m *manager) createInitWithdrawLink(k1 string) (lnurl string, err error) {
	lnurl, err = golnurl.LNURLEncode(fmt.Sprintf(withdrawRequestEndpoint, m.hostname, k1))
	if err != nil {
		return "", err
	}
	return lnurl, nil
}
