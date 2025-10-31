package utils

import "regexp"

var eth1AddressRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{40}$")
var withdrawalCredentialsRE = regexp.MustCompile("^(0x)?00[0-9a-fA-F]{62}$")
var withdrawalCredentialsAddressRE = regexp.MustCompile("^(0x)?010000000000000000000000[0-9a-fA-F]{40}$")
var eth1TxRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{64}$")
var zeroHashRE = regexp.MustCompile("^(0x)?0+$")
var hashRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{96}$")
var HashLikeRegex = regexp.MustCompile(`^[0-9a-fA-F]{0,96}$`)

// IsValidEth1Address verifies whether a string represents a valid eth1-address.
func IsValidEth1Address(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1AddressRE.MatchString(s)
}

// IsEth1Address verifies whether a string represents an eth1-address.
// In contrast to IsValidEth1Address, this also returns true for the 0x0 address
func IsEth1Address(s string) bool {
	return eth1AddressRE.MatchString(s)
}

// IsValidEth1Tx verifies whether a string represents a valid eth1-tx-hash.
func IsValidEth1Tx(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1TxRE.MatchString(s)
}

// IsEth1Tx verifies whether a string represents an eth1-tx-hash.
// In contrast to IsValidEth1Tx, this also returns true for the 0x0 address
func IsEth1Tx(s string) bool {
	return eth1TxRE.MatchString(s)
}

// IsHash verifies whether a string represents an eth1-hash.
func IsHash(s string) bool {
	return hashRE.MatchString(s)
}

// IsValidWithdrawalCredentials verifies whether a string represents valid withdrawal credentials.
func IsValidWithdrawalCredentials(s string) bool {
	return withdrawalCredentialsRE.MatchString(s) || withdrawalCredentialsAddressRE.MatchString(s)
}

// IsValidWithdrawalCredentialsAddress verifies whether a string represents valid withdrawal credential with address.
func IsValidWithdrawalCredentialsAddress(s string) bool {
	return withdrawalCredentialsAddressRE.MatchString(s)
}
