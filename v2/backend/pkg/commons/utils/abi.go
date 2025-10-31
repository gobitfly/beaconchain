package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var ErrRateLimit = errors.New("## RATE LIMIT ##")

func TryFetchContractMetadata(address []byte) (*types.ContractMetadata, error) {
	return getABIFromEtherscan(address)
}

func GetEtherscanAPIBaseUrl(provideDefault bool) string {
	const mainnetBaseUrl = "api.etherscan.io"
	const goerliBaseUrl = "api-goerli.etherscan.io"
	const sepoliaBaseUrl = "api-sepolia.etherscan.io"

	// check config first
	if len(Config.EtherscanAPIBaseURL) > 0 {
		return Config.EtherscanAPIBaseURL
	}

	// check chain id
	switch Config.Chain.ClConfig.DepositChainID {
	case 1: // mainnet
		return mainnetBaseUrl
	case 5: // goerli
		return goerliBaseUrl
	case 11155111: // sepolia
		return sepoliaBaseUrl
	}

	// use default
	if provideDefault {
		return mainnetBaseUrl
	}
	return ""
}

func getABIFromEtherscan(address []byte) (*types.ContractMetadata, error) {
	baseUrl := GetEtherscanAPIBaseUrl(false)
	if len(baseUrl) < 1 {
		return nil, nil
	}

	httpClient := http.Client{Timeout: time.Second * 5}
	resp, err := httpClient.Get(fmt.Sprintf("https://%s/api?module=contract&action=getsourcecode&address=0x%x&apikey=%s", baseUrl, address, Config.EtherscanAPIKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("StatusCode: '%d', Status: '%s'", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	headerData := &struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{}
	err = json.Unmarshal(body, headerData)
	if err != nil {
		return nil, err
	}
	if headerData.Status == "0" {
		if headerData.Message == "NOTOK" {
			return nil, ErrRateLimit
		}
		return nil, fmt.Errorf("%s", headerData.Message)
	}

	data := &types.EtherscanContractMetadata{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}
	if data.Result[0].Abi == "Contract source code not verified" {
		return nil, nil
	}

	contractAbi, err := abi.JSON(strings.NewReader(data.Result[0].Abi))
	if err != nil {
		return nil, err
	}
	meta := &types.ContractMetadata{}
	meta.ABIJson = []byte(data.Result[0].Abi)
	meta.ABI = &contractAbi
	meta.Name = data.Result[0].ContractName
	return meta, nil
}
