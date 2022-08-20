package auth

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

type HeaderApiKey int64

const (
	Admin HeaderApiKey = iota
	User
	Wallet
	Invalid
)

var (
	UsernameHeader     = "x-username"
	AdminApiKeyHeader  = "x-admin-api-key"
	UserApiKeyHeader   = "x-user-api-key"
	WalletApiKeyHeader = "x-wallet-api-key"
	apiKeyHeaders      = map[string]HeaderApiKey{
		AdminApiKeyHeader:  Admin,
		UserApiKeyHeader:   User,
		WalletApiKeyHeader: Wallet,
	}
)

func getApiKey(md metadata.MD) (string, HeaderApiKey, error) {
	numHeaders := 0
	var (
		key     string
		keyType HeaderApiKey
	)

	for header, headerType := range apiKeyHeaders {
		currentKey, err := getStringHeader(md, header)
		if err == nil {
			numHeaders += 1
			key = currentKey
			keyType = headerType
		}
	}

	if numHeaders != 1 {
		return "", Invalid, ErrInvalidHeaderFormat
	}
	return key, keyType, nil
}

func getIntHeader(md metadata.MD, headerName string) (uint64, error) {
	if len(md[headerName]) < 1 {
		return 0, ErrMissingUsername
	}
	id, err := strconv.ParseUint(md[headerName][0], 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getStringHeader(md metadata.MD, headerName string) (string, error) {
	if len(md[headerName]) < 1 {
		return "", ErrInvalidHeaderFormat
	}
	return md[headerName][0], nil
}

func getMetadata(ctx context.Context) (metadata.MD, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrParsingContext
	} else {
		return md, nil
	}
}
