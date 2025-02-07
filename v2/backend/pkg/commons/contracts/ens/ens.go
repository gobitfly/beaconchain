package ens

var (
	registryABI, _               = ENSRegistryMetaData.GetAbi()
	registrarControllerABI, _    = ENSETHRegistrarControllerMetaData.GetAbi()
	oldRegistrarControllerABI, _ = ENSOldRegistrarControllerMetaData.GetAbi()
	publicResolverABI, _         = ENSPublicResolverMetaData.GetAbi()
)

var (
	RegistryNewResolverTopic = registryABI.Events["NewResolver"].ID
	RegistryNewOwnerTopic    = registryABI.Events["NewOwner"].ID
	RegistryNewTTLTopic      = registryABI.Events["NewTTL"].ID
)

var (
	RegistrarControllerNameRegisteredTopic = registrarControllerABI.Events["NameRegistered"].ID
	RegistrarControllerNameRenewedTopic    = registrarControllerABI.Events["NameRenewed"].ID
)

var (
	OldRegistrarControllerNameRegisteredTopic = oldRegistrarControllerABI.Events["NameRegistered"].ID
	OldRegistrarControllerNameRenewedTopic    = oldRegistrarControllerABI.Events["NameRenewed"].ID
)

var (
	PublicResolverNameChangedTopic    = publicResolverABI.Events["NameChanged"].ID
	PublicResolverAddressChangedTopic = publicResolverABI.Events["AddressChanged"].ID
)

var ENSCrontractAddressesEthereum = map[string]string{
	EthereumRegistry:                  "Registry",
	EthereumRegistrarController:       "ETHRegistrarController",
	EthereumOldEnsRegistrarController: "OldEnsRegistrarController",
}

var (
	EthereumRegistry                  = "0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e"
	EthereumRegistrarController       = "0x253553366Da8546fC250F225fe3d25d0C782303b"
	EthereumOldEnsRegistrarController = "0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5"
)

var ENSCrontractAddressesHolesky = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry",
	"0x179Be112b24Ad4cFC392eF8924DfA08C20Ad8583": "ETHRegistrarController",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController",
}

var ENSCrontractAddressesSepolia = map[string]string{
	"0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e": "Registry",
	"0xFED6a969AaA60E4961FCD3EBF1A2e8913ac65B72": "ETHRegistrarController",
	"0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5": "OldEnsRegistrarController",
}

func ENSContractFor(chainID string) map[string]string {
	switch chainID {
	case "1":
		return ENSCrontractAddressesEthereum
	case "17000":
		return ENSCrontractAddressesHolesky
	case "11155111":
		return ENSCrontractAddressesSepolia
	default:
		return nil
	}
}
