package util

import (
	"errors"
	"regexp"
)

var (
	// any non-empty combination of latin character, arabic digit, hyphen, at sign, or period of at max 64 length
	walletIdRegex = regexp.MustCompile(`^[\w-@.]{1,64}$`)
	// limit whitespace to space & each space must be surrounded by non-whitespace character
	walletNameRegex = regexp.MustCompile(`^(\S+( \S+)*)*$`)
	// any latin character or arabic digit, with the option of hyphens and periods if
	// surrounded by latin character or arabic digits.
	usernameRegex = regexp.MustCompile(`^[a-zA-Z\d]+((.|-)[a-zA-Z\d]+)*$`)
)

func ValidateWalletID(id string) error {
	if len(id) == 0 {
		return errors.New("wallet id cannot be blank")
	} else if len(id) > 64 {
		return errors.New("wallet id must be at most 64 characters")
	} else if walletIdRegex.MatchString(id) {
		return nil
	} else {
		return errors.New("wallet id must contain only characters in (a-z,A-Z,0-9,_,-,@,.)")
	}
}

func ValidateWalletName(name string) error {
	if len(name) > 64 {
		return errors.New("wallet name must be at most 64 characters")
	} else if walletNameRegex.MatchString(name) {
		return nil
	} else {
		return errors.New("wallet name must contain only spaces and non-whitespace characters. " +
			"Each space must be surrounded by non-whitespaces.")
	}
}

func ValidateUsername(name string) error {
	if name == "" {
		return errors.New("username cannot be blank")
	} else if len(name) > 64 {
		return errors.New("username must be at most 64 characters")
	} else if usernameRegex.MatchString(name) {
		return nil
	} else {
		return errors.New("username must be alphanumeric. " +
			"If it contains hyphens or periods it must be surrounded by alphanumeric characters")
	}
}
