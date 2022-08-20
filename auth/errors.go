package auth

import "errors"

var (
	MsgAdminAuthorized         = "admin successfully authorized"
	MsgFailedAuthentication    = "failed to authenticate"
	MsgInvalidHeaderFormat     = "invalid header format"
	MsgMissingUsername         = "missing x-username"
	MsgMissingUsernameOrWallet = "missing x-username or wallet"
	ErrUnauthenticated         = errors.New("authentication failed")
	ErrInvalidHeaderFormat     = errors.New(MsgInvalidHeaderFormat)
	ErrParsingContext          = errors.New("unable to get metadata from context")
	ErrMissingUsername         = errors.New(MsgMissingUsername)
	ErrMissingUsernameOrWallet = errors.New(MsgMissingUsernameOrWallet)
	ErrInternal                = errors.New("internal error")
)
