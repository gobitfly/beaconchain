package data_test

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
)

var startTime, _ = time.Parse(time.DateTime, "2024-11-28 00:00:11")
var midTime, _ = time.Parse(time.DateTime, "2024-11-28 12:00:00")
var endTime, _ = time.Parse(time.DateTime, "2024-11-28 23:59:59")

var oneAddress = hundredAddresses[0:1]

var twoAddresses = hundredAddresses[0:2]

var fiveAddresses = hundredAddresses[0:5]

var tenAddresses = hundredAddresses[0:10]

var twentyAddresses = hundredAddresses[0:20]

var fiftyAddresses = hundredAddresses[0:50]

var hundredAddresses = []common.Address{
	common.HexToAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"),
	common.HexToAddress("0x388C818CA8B9251b393131C08a736A67ccB19297"),
	common.HexToAddress("0x9C19B0497997Fe9E75862688a295168070456951"),
	common.HexToAddress("0xf3B0073E3a7F747C7A38B36B805247B222C302A3"),
	common.HexToAddress("0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"),
	common.HexToAddress("0x9AA65464b4cFbe3Dc2BDB3dF412AeE2B3De86687"),
	common.HexToAddress("0x1f9090aaE28b8a3dCeaDf281B0F12828e676c326"),
	common.HexToAddress("0xDd9fd6b6F8f7ea932997992bbE67EabB3e316f3C"),
	common.HexToAddress("0xb8bA36E591FAceE901FfD3d5D82dF491551AD7eF"),
	common.HexToAddress("0x00000000219ab540356cBB839Cbe05303d7705Fa"),
	common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
	common.HexToAddress("0xA9D1e08C7793af67e9d92fe308d5697FB81d3E43"),
	common.HexToAddress("0x49048044D57e1C92A77f79988d21Fa8fAF74E97e"),
	common.HexToAddress("0x1Db92e2EeBC8E0c075a02BeA49a2935BcD2dFCF4"),
	common.HexToAddress("0x28C6c06298d514Db089934071355E5743bf21d60"),
	common.HexToAddress("0x267be1C1D684F78cb4F6a176C4911b741E4Ffdc0"),
	common.HexToAddress("0xd19d4B5d358258f05D7B411E21A1460D11B0876F"),
	common.HexToAddress("0xf7858Da8a6617f7C6d0fF2bcAFDb6D2eeDF64840"),
	common.HexToAddress("0xA7EFAe728D2936e78BDA97dc267687568dD593f3"),
	common.HexToAddress("0xf89d7b9c864f589bbF53a82105107622B35EaA40"),
	common.HexToAddress("0x308861A430be4cce5502d0A12724771Fc6DaF216"),
	common.HexToAddress("0x3999D2c5207C06BBC5cf8A6bEa52966cabB76d41"),
	common.HexToAddress("0xCBD6832Ebc203e49E2B771897067fce3c58575ac"),
	common.HexToAddress("0xa30D8157911ef23c46C0eB71889eFe6a648a41F7"),
	common.HexToAddress("0x1C727a55eA3c11B0ab7D3a361Fe0F3C47cE6de5d"),
	common.HexToAddress("0xBf94F0AC752C739F623C463b5210a7fb2cbb420B"),
	common.HexToAddress("0xae0Ee0A63A2cE6BaeEFFE56e7714FB4EFE48D419"),
	common.HexToAddress("0xe3031C1BfaA7825813c562CbDCC69d96FCad2087"),
	common.HexToAddress("0x45300136662dD4e58fc0DF61E6290DFfD992B785"),
	common.HexToAddress("0x167cB3F2446F829eb327344b66E271D1a7eFeC9A"),
	common.HexToAddress("0x6081258689a75d253d87cE902A8de3887239Fe80"),
	common.HexToAddress("0x55FE002aefF02F77364de339a1292923A15844B8"),
	common.HexToAddress("0x6774Bcbd5ceCeF1336b5300fb5186a12DDD8b367"),
	common.HexToAddress("0x9642b23Ed1E01Df1092B92641051881a322F5D4E"),
	common.HexToAddress("0xA62142888ABa8370742bE823c1782D17A0389Da1"),
	common.HexToAddress("0xFF1F2B4ADb9dF6FC8eAFecDcbF96A2B351680455"),
	common.HexToAddress("0x03cb0021808442Ad5EFb61197966aef72a1deF96"),
	common.HexToAddress("0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5"),
	common.HexToAddress("0x167A9333BF582556f35Bd4d16A7E80E191aa6476"),
	common.HexToAddress("0x19b5cc75846BF6286d599ec116536a333C4C2c14"),
	common.HexToAddress("0xB8001C3eC9AA1985f6c747E25c28324E4A361ec1"),
	common.HexToAddress("0x283Af0B28c62C092C9727F1Ee09c02CA627EB7F5"),
	common.HexToAddress("0x64192819Ac13Ef72bF6b5AE239AC672B43a9AF08"),
	common.HexToAddress("0x4976A4A02f38326660D17bf34b431dC6e2eb2327"),
	common.HexToAddress("0xF19308F923582A6f7c465e5CE7a9Dc1BEC6665B1"),
	common.HexToAddress("0xeBec795c9c8bBD61FFc14A6662944748F299cAcf"),
	common.HexToAddress("0x787B8840100d9BaAdD7463f4a73b5BA73B00C6cA"),
	common.HexToAddress("0xDFd5293D8e347dFe59E90eFd55b2956a1343963d"),
	common.HexToAddress("0x21a31Ee1afC51d94C2eFcCAa2092aD1028285549"),
	common.HexToAddress("0xf584F8728B874a6a5c7A8d4d387C9aae9172D621"),
	common.HexToAddress("0x3CC936b795A188F0e246cBB2D74C5Bd190aeCF18"),
	common.HexToAddress("0x5d22045DAcEAB03B158031eCB7D9d06Fad24609b"),
	common.HexToAddress("0x9696f59E4d72E237BE84fFD425DCaD154Bf96976"),
	common.HexToAddress("0xa26e73C8E9507D50bF808B7A2CA9D5dE4fcC4A04"),
	common.HexToAddress("0x56Eddb7aa87536c09CCc2793473599fD21A8b17F"),
	common.HexToAddress("0xb5d85CBf7cB3EE0D56b3bB207D5Fc4B82f43F511"),
	common.HexToAddress("0x2a0c0DBEcC7E4D658f48E01e3fA353F44050c208"),
	common.HexToAddress("0xCFFAd3200574698b78f32232aa9D63eABD290703"),
	common.HexToAddress("0xDFaa75323fB721e5f29D43859390f62Cc4B600b8"),
	common.HexToAddress("0x8d12A197cB00D4747a1fe03395095ce2A5CC6819"),
	common.HexToAddress("0xC12C41bcd9448E42f16C09f05d99E58C73B784a0"),
	common.HexToAddress("0xaBEA9132b05A70803a4E85094fD0e1800777fBEF"),
	common.HexToAddress("0xe688b84b23f322a994A53dbF8E15FA82CDB71127"),
	common.HexToAddress("0x0000000000A39bb272e79075ade125fd351887Ac"),
	common.HexToAddress("0x0000000000000000000000000000000000000000"),
	common.HexToAddress("0x253553366Da8546fC250F225fe3d25d0C782303b"),
	common.HexToAddress("0x264bd8291fAE1D75DB2c5F573b07faA6715997B5"),
	common.HexToAddress("0x0D0707963952f2fBA59dD06f2b425ace40b492Fe"),
	common.HexToAddress("0x48EC5560bFD59b95859965cCE48cC244CFDF6b0c"),
	common.HexToAddress("0x6262998Ced04146fA42253a5C0AF90CA02dfd2A3"),
	common.HexToAddress("0x46340b20830761efd32832A74d7169B29FEB9758"),
	common.HexToAddress("0xae7ab96520DE3A18E5e111B5EaAb095312D7fE84"),
	common.HexToAddress("0x477b8D5eF7C2C42DB84deB555419cd817c336b6F"),
	common.HexToAddress("0x2a3DD3EB832aF982ec71669E178424b10Dca2EDe"),
	common.HexToAddress("0x77134cbC06cB00b66F4c7e623D5fdBF6777635EC"),
	common.HexToAddress("0xfa9f7a1cBfBCB688729c522b4F0905CcF4d26D25"),
	common.HexToAddress("0xa09bb7afD552A38784A64C8f1e2Dcd0FFe21FFAA"),
	common.HexToAddress("0xC94eBB328aC25b95DB0E0AA968371885Fa516215"),
	common.HexToAddress("0xB3775fB83F7D12A36E0475aBdD1FCA35c091efBe"),
	common.HexToAddress("0xb23360CCDd9Ed1b15D45E5d3824Bb409C8D7c460"),
	common.HexToAddress("0xDBF5E9c5206d0dB70a90108bf936DA60221dC080"),
	common.HexToAddress("0xDB044B8298E04D442FdBE5ce01B8cc8F77130e33"),
	common.HexToAddress("0x6fC48dE8f167456b7aa27dD4ecfaBBA329EA623D"),
	common.HexToAddress("0x1AB4973a48dc892Cd9971ECE8e01DcC7688f8F23"),
	common.HexToAddress("0x1a0ad011913A150f69f6A19DF447A0CfD9551054"),
	common.HexToAddress("0x2b3FeD49557bd88f78b898684F82FBb355305DbB"),
	common.HexToAddress("0x1cE7AE555139c5EF5A57CC8d814a867ee6Ee33D8"),
	common.HexToAddress("0x8D1f2eBFACCf1136dB76FDD1b86f1deDE2D23852"),
	common.HexToAddress("0xEFb2E870b14D7e555a31B392541ACf002Dae6aE9"),
	common.HexToAddress("0xd2674dA94285660c9b2353131bef2d8211369A4B"),
	common.HexToAddress("0xF5C9F957705bea56a7e806943f98F7777B995826"),
	common.HexToAddress("0xD9A442856C234a39a81a089C06451EBAa4306a72"),
	common.HexToAddress("0x974CaA59e49682CdA0AD2bbe82983419A2ECC400"),
	common.HexToAddress("0xF5bEC430576fF1b82e44DDB5a1C93F6F9d0884f3"),
	common.HexToAddress("0xb8901acB165ed027E32754E0FFe830802919727f"),
	common.HexToAddress("0x5f65f7b609678448494De4C87521CdF6cEf1e932"),
	common.HexToAddress("0x5FDCCA53617f4d2b9134B29090C87D01058e27e9"),
	common.HexToAddress("0x416299AAde6443e6F6e8ab67126e65a7F606eeF5"),
	common.HexToAddress("0xf60c2Ea62EDBfE808163751DD0d8693DCb30019c"),
	common.HexToAddress("0xb47e3cd837dDF8e4c57F05d70Ab865de6e193BBB"),
}

func TestSuite(t *testing.T) {
	db := dbFromEnv(t, data.Table)
	store := data.NewStore(db)

	limit := int64(50)
	tests := []struct {
		name      string
		addresses []common.Address
	}{
		{
			name:      "oneAddress",
			addresses: oneAddress,
		},
		{
			name:      "twoAddresses",
			addresses: twoAddresses,
		},
		{
			name:      "fiveAddresses",
			addresses: fiveAddresses,
		},
		{
			name:      "fiftyAddresses",
			addresses: fiftyAddresses,
		},
		{
			name:      "hundredAddresses",
			addresses: hundredAddresses,
		},
	}
	for _, tt := range tests {
		conditions := makeConditions()
		// times := make(map[string]time.Duration)
		t.Run(tt.name, func(t *testing.T) {
			for name, opts := range conditions {
				t.Run(name, func(t *testing.T) {
					// start := time.Now()
					_, _, err := store.Get(tt.addresses, nil, limit, opts...)
					if err != nil {
						t.Fatal(err)
					}
					// times[name] = time.Since(start)
				})
			}
		})
		// fmt.Println(times)
	}
}

var baseConditions = map[string]data.Option{
	"network":  data.ByChainID("1"),
	"date":     data.WithTimeRange(timestamppb.New(startTime), timestamppb.New(midTime)),
	"received": data.OnlyReceived(),
	"sent":     data.OnlySent(),
	"tx":       data.OnlyTransactions(),
	"erc20":    data.OnlyTransfers(),
	"method":   data.ByMethod("a9059cbb"),
	"asset":    data.ByAsset(common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")),
}

func makeConditions() map[string][]data.Option {
	keys := maps.Keys(baseConditions)
	conditionsList := generateCombinations(keys)
	conditions := make(map[string][]data.Option)
	for _, condition := range conditionsList {
		if slices.Contains(condition, "tx") {
			if slices.Contains(condition, "erc20") || slices.Contains(condition, "asset") {
				continue
			}
		}
		if slices.Contains(condition, "erc20") {
			if slices.Contains(condition, "tx") || slices.Contains(condition, "method") {
				continue
			}
		}
		if slices.Contains(condition, "method") {
			if slices.Contains(condition, "erc20") || slices.Contains(condition, "asset") {
				continue
			}
		}
		if slices.Contains(condition, "asset") {
			if slices.Contains(condition, "tx") || slices.Contains(condition, "method") {
				continue
			}
		}
		if slices.Contains(condition, "sent") && slices.Contains(condition, "received") {
			continue
		}
		var opts []data.Option
		for _, key := range condition {
			opts = append(opts, baseConditions[key])
		}
		conditions[strings.Join(condition, " ")] = opts
	}
	conditions["no filter"] = []data.Option{data.WithNoOption()}
	return conditions
}

func generateCombinations(elements []string) [][]string {
	n := len(elements)
	numCombinations := int(math.Pow(2, float64(n)))
	var result [][]string

	// Iterate through all possible combinations using bitmasks
	for i := 1; i < numCombinations; i++ {
		var combination []string
		for j := 0; j < n; j++ {
			// Check if j-th element is included in the current combination
			if i&(1<<j) != 0 {
				combination = append(combination, elements[j])
			}
		}
		result = append(result, combination)
	}

	return result
}
