package auth

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/resources/user"
	"github.com/xbit-gg/xln/resources/wallet"
)

type Service interface {
	GetIdentityOfApiKey(apiKey string) (*IdentityType, error)

	// ValidateAdminCredentials reads credentials from the request context and compares them to admin API key.
	// Takes the name of the selector and logs it.
	// An error is returned if the credentials do not match.
	ValidateAdminCredentials(ctx context.Context, selector string) error
	// ValidateUserCredentials reads credentials from the request context and compares them to the API key of the user.
	// The credentials are also compared with the admin credentials.
	// Takes the name of the selector and logs it.
	// An error is returned if the credentials do not match.
	ValidateUserCredentials(ctx context.Context, selector string) (username string, err error)
	// ValidateWalletCredentials reads creds from the request context and compares them to the API key of the wallet id.
	// The credentials are also compared with the admin and user credentials.
	// Takes the name of the selector and logs it.
	// An error is returned if the credentials do not match.
	ValidateWalletCredentials(ctx context.Context, walletId string, selector string) (username string, err error)
}

type service struct {
	xlnApiKey string
	users     *user.Manager
	wallets   *wallet.Manager

	userKeys       *cache.Cache
	walletKeys     *cache.Cache
	walletUserKeys *cache.Cache
}

// NewService returns a new authentication service.
func NewService(xlnApiKey string, userManager *user.Manager, walletManager *wallet.Manager) Service {
	return &service{
		xlnApiKey:      xlnApiKey,
		users:          userManager,
		wallets:        walletManager,
		userKeys:       cache.New(12*time.Hour, 24*time.Hour),
		walletKeys:     cache.New(12*time.Hour, 24*time.Hour),
		walletUserKeys: cache.New(12*time.Hour, 24*time.Hour),
	}
}

type IdentityType struct {
	Admin  bool
	User   string
	Wallet string
}

func (s *service) GetIdentityOfApiKey(apiKey string) (*IdentityType, error) {
	// check if admin key
	if s.xlnApiKey == apiKey {
		log.Info("admin requested information on its apikey privileges")
		return &IdentityType{Admin: true}, nil
	}
	// check if user key
	if user, err := (*s.users).GetUserWithApiKey(apiKey); err == models.ErrUserNotFound {
		// then check if api key belongs to wallet
		if wallet, err := (*s.wallets).GetWalletWithApiKey(apiKey); err == models.ErrWalletNotFound {
			// key doesn't belong to any resource.
			return nil, nil
		} else if err != nil {
			return nil, err
		} else if wallet != nil {
			return &IdentityType{Wallet: wallet.ID}, nil
		} else {
			log.WithError(ErrInternal).Error("getting identity of api key failed. Reason: expected repo to return wallet and no error or no wallet and an error. Instead nil and nil were given.")
			return nil, ErrInternal
		}
	} else if err != nil {
		return nil, err
	} else if user != nil {
		return &IdentityType{User: user.Username}, nil
	} else {
		log.WithError(ErrInternal).Error("getting identity of api key failed. Reason: expected user and no error, or no user and an error. Instead nil and nil were given.")
		return nil, ErrInternal
	}
}

func (s *service) ValidateAdminCredentials(ctx context.Context, selector string) error {
	md, err := getMetadata(ctx)
	if err != nil {
		return err
	}
	apiKey, err := getStringHeader(md, AdminApiKeyHeader)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"selector": selector,
		}).Warn("failed to authenticate admin")
		return err
	}
	if s.xlnApiKey == apiKey {
		return nil
	} else {
		log.WithError(ErrUnauthenticated).WithFields(log.Fields{
			"selector": selector,
			"apiKey":   apiKey,
		}).Warn("failed to authenticate admin. Reason: unrecognized api key was used")
		return ErrUnauthenticated
	}
}

func (s *service) ValidateUserCredentials(ctx context.Context, selector string) (string, error) {
	md, err := getMetadata(ctx)
	if err != nil {
		return "", err
	}
	username, err := getStringHeader(md, UsernameHeader)
	if err != nil {
		return "", err
	}
	apiKey, keyType, err := getApiKey(md)
	if err != nil {
		return username, err
	}
	switch keyType {
	case Admin:
		if s.xlnApiKey != apiKey {
			log.WithError(ErrUnauthenticated).WithFields(log.Fields{
				"selector": selector,
				"apiKey":   apiKey,
				"user":     username,
			}).Warn("failed to authenticate admin. Reason: unrecognized api key was used")
			return username, ErrUnauthenticated
		} else if exists, err := s.userExists(username); err == nil && exists {
			log.WithFields(log.Fields{
				"selector": selector,
				"user":     username,
			}).Info(MsgAdminAuthorized)
			return username, nil
		} else {
			return username, err
		}
	case User:
		if s.userKeyMatchesUser(username, apiKey) {
			return username, nil
		}
	default:
		return username, ErrInvalidHeaderFormat
	}
	log.WithError(ErrUnauthenticated).WithFields(log.Fields{
		"selector": selector,
		"user":     username,
	}).Warn("failed to authenticate user")
	return username, ErrUnauthenticated
}

func (s *service) ValidateWalletCredentials(ctx context.Context, walletId string, selector string) (string, error) {
	md, headerErr := getMetadata(ctx)
	if headerErr != nil {
		return "", headerErr
	}
	// get apikey header
	apiKey, keyType, headerErr := getApiKey(md)
	if headerErr != nil {
		return "", headerErr
	}
	// get username header
	username, headerErr := getStringHeader(md, UsernameHeader)
	if headerErr == ErrMissingUsername {
		return username, headerErr
	} else if headerErr != nil {
		return username, headerErr
	}

	switch keyType {
	case Admin:
		if s.xlnApiKey != apiKey {
			log.WithError(ErrUnauthenticated).WithFields(log.Fields{
				"selector": selector,
				"apiKey":   apiKey,
				"user":     username,
				"wallet":   walletId,
			}).Warn("failed to authenticate admin. Reason: unrecognized api key was used")
			return username, ErrUnauthenticated
		} else if exists, err := s.userExists(username); err == nil && exists {
			log.WithFields(log.Fields{
				"selector": selector,
				"user":     username,
				"wallet":   walletId,
			}).Info(MsgAdminAuthorized)
			return username, nil
		} else {
			return username, err
		}
	case User:
		if s.userKeyMatchesUser(username, apiKey) &&
			s.userMatchesWallet(username, walletId, apiKey) {
			return username, nil
		}
	case Wallet:
		if s.walletKeyMatchesWallet(username, walletId, apiKey) {
			return username, nil
		}
	default:
		return username, ErrInvalidHeaderFormat
	}
	log.WithError(ErrUnauthenticated).WithFields(log.Fields{
		"selector": selector,
		"user":     username,
		"wallet":   walletId,
	}).Warn("failed to authenticate wallet")
	return username, ErrUnauthenticated
}

func (s *service) userExists(username string) (bool, error) {
	if username != "" {
		if _, err := (*s.users).GetUser(username); err == models.ErrUserNotFound {
			log.WithField("user", username).WithError(err).Warn("admin call was attempted, but user not found.")
			return false, ErrUnauthenticated
		} else if err != nil {
			log.WithField("user", username).WithError(err).Error("admin call was attempted, but failed")
			return false, ErrInternal
		} else {
			return true, nil
		}
	} else {
		return false, ErrMissingUsername
	}
}

func (s *service) walletExists(username, wallet string) (bool, error) {
	if username != "" && wallet != "" {
		if _, err := (*s.wallets).GetWallet(username, wallet); err == models.ErrWalletNotFound {
			log.WithField("user", username).WithError(err).Warn("admin call was attempted, but wallet not found.")
			return false, ErrUnauthenticated
		} else if err != nil {
			log.WithField("user", username).WithError(err).Error("admin call was attempted, but failed")
			return false, ErrInternal
		} else {
			return true, nil
		}
	} else {
		return false, ErrMissingUsernameOrWallet
	}
}

func (s *service) userKeyMatchesUser(username string, userKey string) bool {
	if userKey == "" {
		return false
	}
	var apiKey string
	if cacheKey, contains := s.userKeys.Get(username); contains {
		apiKey = cacheKey.(string)
	} else {
		log.WithField("user", username).Debug("Cache miss for user authentication")
		u, err := (*s.users).GetUser(username)
		if err != nil {
			return false
		}
		apiKey = u.ApiKey
		s.userKeys.SetDefault(username, apiKey)
	}
	return apiKey == userKey
}

func (s *service) userMatchesWallet(username string, walletId string, userKey string) bool {
	if userKey == "" {
		return false
	}
	if cacheKey, contains := s.walletUserKeys.Get(walletId); contains {
		apiKey := cacheKey.(string)
		return apiKey == userKey
	} else {
		log.WithField("wallet", walletId).Debug("Cache miss for wallet authentication")
		_, err := (*s.wallets).GetWallet(username, walletId)
		if err != nil {
			return false
		} else {
			s.walletUserKeys.SetDefault(walletId, userKey)
			return true
		}
	}
}

func (s *service) walletKeyMatchesWallet(username, walletId, walletKey string) bool {
	if walletKey == "" {
		return false
	}
	var apiKey string
	if cacheKey, contains := s.walletKeys.Get(walletId); contains {
		apiKey = cacheKey.(string)
	} else {
		log.WithField("wallet", walletId).Debug("Cache miss for wallet authentication")
		w, err := (*s.wallets).GetWallet(username, walletId)
		if err != nil {
			return false
		}
		apiKey = w.ApiKey
		s.walletKeys.SetDefault(walletId, apiKey)
	}
	return apiKey == walletKey
}
