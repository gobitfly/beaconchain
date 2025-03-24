package types

import "regexp"

var (
	ReName                         = regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]*$`)
	ReInteger                      = regexp.MustCompile(`^[0-9]+$`)
	ReValidatorDashboardPublicId   = regexp.MustCompile(`^v-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	ReValidatorPublicKeyWithPrefix = regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`)
	ReValidatorPublicKey           = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{96}$`)
	ReValidatorList                = regexp.MustCompile(`^(0x[0-9a-fA-F]{96}|[0-9]+)(,\s*(0x[0-9a-fA-F]{96}|[0-9]+)\s*)+$`)
	ReEthereumAddress              = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{40}$`)
	ReWithdrawalCredential         = regexp.MustCompile(`^(0x)?0[012][0-9a-fA-F]{62}$`)
	ReTransactionHash              = regexp.MustCompile(`^0x[0-9a-fA-F]{62}$`)
	ReEnsName                      = regexp.MustCompile(`^.+\.eth$`)
	ReGraffiti                     = regexp.MustCompile(`^.{2,32}$`) // at least 2 characters, so that queries won't time out
	ReGraffitiHex                  = regexp.MustCompile(`^(0x)?([0-9a-fA-F]{2}){32}$`)
	ReCursor                       = regexp.MustCompile(`^[A-Za-z0-9-_]+$`) // has to be base64
	ReEmail                        = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	RePassword                     = regexp.MustCompile(`^.{5,}$`)
	ReEmailUserToken               = regexp.MustCompile(`^[a-z0-9]{40}$`)
	ReJsonContentType              = regexp.MustCompile(`^application\/json(;.*)?$`)
)
