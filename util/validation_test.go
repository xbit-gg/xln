package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateID(t *testing.T) {
	require.NoError(t, ValidateWalletID("A"))
	require.NoError(t, ValidateWalletID("a"))
	require.NoError(t, ValidateWalletID("0"))
	require.NoError(t, ValidateWalletID("_"))
	require.NoError(t, ValidateWalletID("-"))
	require.NoError(t, ValidateWalletID("."))
	require.NoError(t, ValidateWalletID("@"))
	require.NoError(t, ValidateWalletID("Aa0_-.@"))
	require.NoError(t, ValidateWalletID(strings.Repeat("A", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat("a", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat("0", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat("_", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat("-", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat(".", 64)))
	require.NoError(t, ValidateWalletID(strings.Repeat("@", 64)))

	require.Error(t, ValidateWalletID(" a"), "id cannot begin with space")
	require.Error(t, ValidateWalletID("a "), "id cannot end with space")
	require.Error(t, ValidateWalletID("a  a"), "id cannot contain consecutive spaces")
	require.Error(t, ValidateWalletID("a   a"), "id cannot contain consecutive spaces")

	require.Error(t, ValidateWalletID("a\t"), "id cannot end with whitespace")
	require.Error(t, ValidateWalletID("a\n"), "id cannot end with whitespace")
	require.Error(t, ValidateWalletID("a\f"), "id cannot end with whitespace")
	require.Error(t, ValidateWalletID("a\r"), "id cannot end with whitespace")
	require.Error(t, ValidateWalletID("\ta"), "id cannot begin with whitespace")
	require.Error(t, ValidateWalletID("\na"), "id cannot begin with whitespace")
	require.Error(t, ValidateWalletID("\fa"), "id cannot begin with whitespace")
	require.Error(t, ValidateWalletID("\ra"), "id cannot begin with whitespace")

	require.Error(t, ValidateWalletID(strings.Repeat("a", 65)), "id that is too long should error")
	require.Error(t, ValidateWalletID(strings.Repeat("a", 1000)), "id that is too long should error")
}

func TestValidateWalletName(t *testing.T) {
	require.NoError(t, ValidateWalletName("a"))
	require.NoError(t, ValidateWalletName("A"))
	require.NoError(t, ValidateWalletName("0"))
	require.NoError(t, ValidateWalletName("a a"))
	require.NoError(t, ValidateWalletName("Aa-0 a.aa-é.9-1_A"))
	require.NoError(t, ValidateWalletName("Białołęka"), "wallet name should allow non-English characters")
	require.NoError(t, ValidateWalletName("الخرج"), "wallet name should allow non-English characters")
	require.NoError(t, ValidateWalletName(""), "empty wallet name is valid")
	require.NoError(t, ValidateWalletName(strings.Repeat("a", 64)))

	require.Error(t, ValidateWalletName(" a"), "wallet name cannot begin with space")
	require.Error(t, ValidateWalletName("a "), "wallet name cannot end with space")
	require.Error(t, ValidateWalletName("a  a"), "wallet name cannot contain consecutive spaces")
	require.Error(t, ValidateWalletName("a   a"), "wallet name cannot contain consecutive spaces")

	require.Error(t, ValidateWalletName("a\t"), "wallet name cannot end with whitespace")
	require.Error(t, ValidateWalletName("a\n"), "wallet name cannot end with whitespace")
	require.Error(t, ValidateWalletName("a\f"), "wallet name cannot end with whitespace")
	require.Error(t, ValidateWalletName("a\r"), "wallet name cannot end with whitespace")
	require.Error(t, ValidateWalletName("\ta"), "wallet name cannot begin with whitespace")
	require.Error(t, ValidateWalletName("\na"), "wallet name cannot begin with whitespace")
	require.Error(t, ValidateWalletName("\fa"), "wallet name cannot begin with whitespace")
	require.Error(t, ValidateWalletName("\ra"), "wallet name cannot begin with whitespace")

	require.Error(t, ValidateWalletName(strings.Repeat("a", 65)), "wallet name that is too long should error")
	require.Error(t, ValidateWalletName(strings.Repeat("a", 1000)), "wallet name that is too long should error")
}

func TestValidateUsername(t *testing.T) {
	require.NoError(t, ValidateUsername("A"))
	require.NoError(t, ValidateUsername("a"))
	require.NoError(t, ValidateUsername("0"))
	require.NoError(t, ValidateUsername("a-a"))
	require.NoError(t, ValidateUsername("aa-aa"))
	require.NoError(t, ValidateUsername("a.a"))
	require.NoError(t, ValidateUsername("aa.aa"))
	require.NoError(t, ValidateUsername("0.1"))
	require.NoError(t, ValidateUsername("01.01"))
	require.NoError(t, ValidateUsername("l33t"))
	require.NoError(t, ValidateUsername("u1"))
	require.NoError(t, ValidateUsername("1u"))

	require.Error(t, ValidateUsername(".a"), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername("a."), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername("a-"), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername("-a"), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername("-.a"), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername(".-a"), "dash and period must be surrounded by alphanumeric character")
	require.Error(t, ValidateUsername("é"), "username should only allow alphanumeric characters")
	require.Error(t, ValidateUsername("Białołęka"), "username should only allow alphanumeric characters")
	require.Error(t, ValidateUsername("الخرج"), "username should only allow alphanumeric characters")
	require.Error(t, ValidateUsername("My name!"), "username should only allow alphanumeric characters")
	require.Error(t, ValidateUsername(""), "username cannot be blank")
	require.Error(t, ValidateUsername(strings.Repeat("a", 65)), "username that is too long should error")
	require.Error(t, ValidateUsername(strings.Repeat("a", 1000)), "username that is too long should error")
}
