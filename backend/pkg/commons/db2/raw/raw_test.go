package raw

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/databasetest"
)

func TestRaw(t *testing.T) {
	client, admin := databasetest.NewBigTable(t)

	s, err := database.NewBigTableWithClient(context.Background(), client, admin, Schema)
	if err != nil {
		t.Fatal(err)
	}

	store := Store{
		db:         database.Wrap(s, BlocksRawTable, ""),
		compressor: noOpCompressor{},
	}

	block := testFullBlock
	if err := store.AddBlocks([]FullBlockData{block}); err != nil {
		t.Fatal(err)
	}

	res, err := store.ReadBlockByNumber(block.ChainID, block.BlockNumber)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(res.Block), testBlock; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Receipts), testReceipts; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Traces), testTraces; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := string(res.Uncles), testUncles; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

var testFullBlock = FullBlockData{
	ChainID:          1,
	BlockNumber:      testBlockNumber,
	BlockHash:        common.HexToHash(testBlockHash).Bytes(),
	BlockUnclesCount: 1,
	Block:            []byte(testBlock),
	Receipts:         []byte(testReceipts),
	Traces:           []byte(testTraces),
	Uncles:           []byte(testUncles),
}

var testTwoUnclesFullBlock = FullBlockData{
	ChainID:          1,
	BlockNumber:      testTwoUnclesBlockNumber,
	BlockUnclesCount: 2,
	Block:            []byte(testTwoUnclesBlock),
	Receipts:         nil,
	Traces:           nil,
	Uncles:           []byte(testTwoUnclesBlockUncles),
}

const (
	testBlockNumber = 6008149
	testBlockHash   = "0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35"
	testBlock       = `{
  "jsonrpc":"2.0",
  "id":1,
  "result":{
    "difficulty":"0xbfabcdbd93dda",
    "extraData":"0x737061726b706f6f6c2d636e2d6e6f64652d3132",
    "gasLimit":"0x79f39e",
    "gasUsed":"0x79ccd3",
    "hash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
    "logsBloom":"0x4848112002a2020aaa0812180045840210020005281600c80104264300080008000491220144461026015300100000128005018401002090a824a4150015410020140400d808440106689b29d0280b1005200007480ca950b15b010908814e01911000054202a020b05880b914642a0000300003010044044082075290283516be82504082003008c4d8d14462a8800c2990c88002a030140180036c220205201860402001014040180002006860810ec0a1100a14144148408118608200060461821802c081000042d0810104a8004510020211c088200420822a082040e10104c00d010064004c122692020c408a1aa2348020445403814002c800888208b1",
    "miner":"0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c",
    "mixHash":"0x3d1fdd16f15aeab72e7db1013b9f034ee33641d92f71c0736beab4e67d34c7a7",
    "nonce":"0x4db7a1c01d8a8072",
    "number":"0x5bad55",
    "parentHash":"0x61a8ad530a8a43e3583f8ec163f773ad370329b2375d66433eb82f005e1d6202",
    "receiptsRoot":"0x5eced534b3d84d3d732ddbc714f5fd51d98a941b28182b6efe6df3a0fe90004b",
    "sha3Uncles":"0x8a562e7634774d3e3a36698ac4915e37fc84a2cd0044cb84fa5d80263d2af4f6",
    "size":"0x41c7",
    "stateRoot":"0xf5208fffa2ba5a3f3a2f64ebd5ca3d098978bedd75f335f56b705d8715ee2305",
    "timestamp":"0x5b541449",
    "totalDifficulty":"0x12ac11391a2f3872fcd",
    "transactions":[
      {
        "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
        "blockNumber":"0x5bad55",
        "from":"0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
        "gas":"0x249f0",
        "gasPrice":"0x174876e800",
        "hash":"0x8784d99762bccd03b2086eabccee0d77f14d05463281e121a62abfebcf0d2d5f",
        "input":"0x6ea056a9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
        "nonce":"0x5e4724",
        "to":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
        "transactionIndex":"0x0",
        "value":"0x0",
        "type":"0x0",
        "chainId":"0x1",
        "v":"0x25",
        "r":"0xd1556332df97e3bd911068651cfad6f975a30381f4ff3a55df7ab3512c78b9ec",
        "s":"0x66b51cbb10cd1b2a09aaff137d9f6d4255bf73cb7702b666ebd5af502ffa4410"
      },
      {
        "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
        "blockNumber":"0x5bad55",
        "from":"0xc837f51a0efa33f8eca03570e3d01a4b2cf97ffd",
        "gas":"0x15f90",
        "gasPrice":"0x14b8d03a00",
        "hash":"0x311be6a9b58748717ac0f70eb801d29973661aaf1365960d159e4ec4f4aa2d7f",
        "input":"0x",
        "nonce":"0x4241",
        "to":"0xf49bd0367d830850456d2259da366a054038dc46",
        "transactionIndex":"0x1",
        "value":"0x1bafa9ee16e78000",
        "type":"0x0",
        "chainId":"0x1",
        "v":"0x25",
        "r":"0xe9ef2f6fcff76e45fac6c2e8080094370082cfb47e8fde0709312f9aa3ec06ad",
        "s":"0x421ebc4ebe187c173f13b1479986dcbff5c4997c0dfeb1fd149a982ad4bcdfe7"
      }
    ],
    "transactionsRoot":"0xf98631e290e88f58a46b7032f025969039aa9b5696498efc76baf436fa69b262",
    "uncles":[
      "0x824cce7c7c2ec6874b9fa9a9a898eb5f27cbaf3991dfa81084c3af60d1db618c"
    ]
  }
}`
	testTraces = `{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "result":{
        "from":"0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
        "gas":"0x249f0",
        "gasUsed":"0xc349",
        "to":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
        "input":"0x6ea056a9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
        "output":"0x0000000000000000000000000000000000000000000000000000000000000001",
        "calls":[
          {
            "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
            "gas":"0x1e530",
            "gasUsed":"0x41b",
            "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
            "input":"0x3c18d3180000000000000000000000000000000000000000000000000000000000000000",
            "output":"0x000000000000000000000000b2233fcec42c588ee71a594d9a25aa695345426c",
            "value":"0x0",
            "type":"CALL"
          },
          {
            "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
            "gas":"0x1ddd2",
            "gasUsed":"0x5e56",
            "to":"0xb2233fcec42c588ee71a594d9a25aa695345426c",
            "input":"0x6ea056a9000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
            "output":"0x0000000000000000000000000000000000000000000000000000000000000001",
            "calls":[
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x1cdaf",
                "gasUsed":"0x2be",
                "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
                "input":"0x97dc97cb",
                "output":"0x0000000000000000000000006cace0528324a8afc2b157ceba3cdd2a27c4e21f",
                "value":"0x0",
                "type":"CALL"
              },
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x1c2d7",
                "gasUsed":"0x2a8",
                "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
                "input":"0x8da5cb5b",
                "output":"0x000000000000000000000000fbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
                "value":"0x0",
                "type":"CALL"
              },
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x1b7d0",
                "gasUsed":"0x294",
                "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
                "input":"0xb9b8af0b",
                "output":"0x0000000000000000000000000000000000000000000000000000000000000000",
                "value":"0x0",
                "type":"CALL"
              },
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x1ae06",
                "gasUsed":"0x300",
                "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
                "input":"0xb269681d",
                "output":"0x000000000000000000000000fbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
                "value":"0x0",
                "type":"CALL"
              },
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x8fc",
                "gasUsed":"0x0",
                "to":"0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
                "input":"0x",
                "value":"0xbd8d7fa6f8cc00",
                "type":"CALL"
              },
              {
                "from":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
                "gas":"0x1844a",
                "gasUsed":"0xa85",
                "to":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
                "input":"0x28090abb0000000000000000000000004b9c25ca0224aef6a7522cabdbc3b2e125b7ca50000000000000000000000000fbb1b73c4f0bda4f67dca266ce6ef42f520fbb98000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
                "value":"0x0",
                "type":"CALL"
              }
            ],
            "type":"DELEGATECALL"
          }
        ],
        "value":"0x0",
        "type":"CALL"
      }
    },
    {
      "result":{
        "from":"0xc837f51a0efa33f8eca03570e3d01a4b2cf97ffd",
        "gas":"0x15f90",
        "gasUsed":"0x5208",
        "to":"0xf49bd0367d830850456d2259da366a054038dc46",
        "input":"0x",
        "value":"0x1bafa9ee16e78000",
        "type":"CALL"
      }
    }
  ]
}`
	testReceipts = `{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
      "blockNumber":"0x5bad55",
      "contractAddress":null,
      "cumulativeGasUsed":"0xc349",
      "effectiveGasPrice":"0x174876e800",
      "from":"0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
      "gasUsed":"0xc349",
      "logs":[
        {
          "address":"0xa3c1e324ca1ce40db73ed6026c4a177f099b5770",
          "topics":[
            "0xa64da754fccf55aa65a1f0128a648633fade3884b236e879ee9f64c78df5d5d7",
            "0x0000000000000000000000004b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
            "0x000000000000000000000000fbb1b73c4f0bda4f67dca266ce6ef42f520fbb98",
            "0x0000000000000000000000000000000000000000000000000000000000000000"
          ],
          "data":"0x00000000000000000000000000000000000000000000000000bd8d7fa6f8cc00",
          "blockNumber":"0x5bad55",
          "transactionHash":"0x8784d99762bccd03b2086eabccee0d77f14d05463281e121a62abfebcf0d2d5f",
          "transactionIndex":"0x0",
          "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
          "logIndex":"0x0",
          "removed":false
        }
      ],
      "logsBloom":"0x00000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000200000000000000800000000000000000000000000800040004000000000000000000000000000000000000000000000000000000020000000000000000000800000000000000000000000100000000000000000000000000800000000000000000000000008000000000000000000000000000000000004000000000400000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000020000000000000000000000000000000000000000000000000000000000000000000",
      "status":"0x1",
      "to":"0x4b9c25ca0224aef6a7522cabdbc3b2e125b7ca50",
      "transactionHash":"0x8784d99762bccd03b2086eabccee0d77f14d05463281e121a62abfebcf0d2d5f",
      "transactionIndex":"0x0",
      "type":"0x0"
    },
    {
      "blockHash":"0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
      "blockNumber":"0x5bad55",
      "contractAddress":null,
      "cumulativeGasUsed":"0x11551",
      "effectiveGasPrice":"0x14b8d03a00",
      "from":"0xc837f51a0efa33f8eca03570e3d01a4b2cf97ffd",
      "gasUsed":"0x5208",
      "logs":[

      ],
      "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "status":"0x1",
      "to":"0xf49bd0367d830850456d2259da366a054038dc46",
      "transactionHash":"0x311be6a9b58748717ac0f70eb801d29973661aaf1365960d159e4ec4f4aa2d7f",
      "transactionIndex":"0x1",
      "type":"0x0"
    }
  ]
}`
	testUncles = `[
  {
    "jsonrpc":"2.0",
    "id":1,
    "result":{
      "difficulty":"0xbf93da424b943",
      "extraData":"0x65746865726d696e652d657539",
      "gasLimit":"0x7a121d",
      "gasUsed":"0x79ea62",
      "hash":"0x824cce7c7c2ec6874b9fa9a9a898eb5f27cbaf3991dfa81084c3af60d1db618c",
      "logsBloom":"0x0948432021200401804810002000000000381001001202440000010020000080a016262050e44850268052000400100505022305a64000054004200b0c04110000080c1055c42001054b804940a0401401008a00112d80082113400c10006580140005011a40220020000010001c0a00082300434002000050840010102082801c2000148540201004491814020480080111a0300600000003800640024200109c00202010044000880000106810a1a010000028a0100000422000140011000050a2a44b3080001060800000540c108102102600d000004730404a880100600021080100403000180000062642408b440060590400080101e046f08000000430",
      "miner":"0xea674fdde714fd979de3edf0f56aa9716b898ec8",
      "mixHash":"0x0b15fe0a9aa789c167b0f5ade7b72969d9f2193014cb4e98382254f60ffb2f4a",
      "nonce":"0xa212d6400b89b3f6",
      "number":"0x5bad54",
      "parentHash":"0x05e19fb68d9ec798073808e8b3170875cb327d4b6cde7d6f60fe194677bb26fd",
      "receiptsRoot":"0x90807b32c4aa4610c57289de57fa68ba50ed53f14dd2c25f1862aa049029dcd6",
      "sha3Uncles":"0xf763576c1ea6a8c61a206e16b1a2451bec5cba1c7545d7ff733a1e8c78715569",
      "size":"0x216",
      "stateRoot":"0xebc7a1603bfffe0a14bdb89f898e2f2824abb40f04579beb7b920c56d6e273c9",
      "timestamp":"0x5b54143f",
      "totalDifficulty":"0x12ac11391a2f3872fcd",
      "transactions":[

      ],
      "transactionsRoot":"0x7562cba41e067b364b933e7b566fb2444f6954fef3964a5a487d4cd79d97a56c",
      "uncles":[

      ]
    }
  }
]`

	testTwoUnclesBlockNumber = 141
	testTwoUnclesBlock       = `{
   "jsonrpc":"2.0",
   "id":0,
   "result":{
      "difficulty":"0x4417decf7",
      "extraData":"0x426974636f696e2069732054484520426c6f636b636861696e2e",
      "gasLimit":"0x1388",
      "gasUsed":"0x0",
      "hash":"0xeafbe76fdcadc1b69ba248589eb2a674b60b00c84374c149c9deaf5596183932",
      "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "miner":"0x1b7047b4338acf65be94c1a3e8c5c9338ad7d67c",
      "mixHash":"0x21eabda67c3151855389a5a968e50daa7b356b3046e2f119ef46c97d204a541e",
      "nonce":"0x85378a3fc5e608e1",
      "number":"0x8d",
      "parentHash":"0xe2c1e8200ef2e9fba09979f0b504dc52c068719623c7064904c7bd3e9365acc1",
      "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "sha3Uncles":"0x393f5f01182846b91386f8b00759fd54f83998a6a1064b8ac72fc8eca1bcf81b",
      "size":"0x653",
      "stateRoot":"0x3e1eea9a01178945535230b6f5839201f594d9be20618bb4edaa383f4f0c850f",
      "timestamp":"0x55ba4444",
      "totalDifficulty":"0x24826e73469",
      "transactions":[
         
      ],
      "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "uncles":[
         "0x61beeeb3e11e89d19fed2e988c8017b55c3ddb8895f531072363ce2abaf56b95",
         "0xf84d9d74415364c3a7569f315ff831b910968c7dd637fffaab51278c9e7f9306"
      ]
   }
}`
	testTwoUnclesBlockUncles = `[
   {
      "jsonrpc":"2.0",
      "id":141,
      "result":{
         "difficulty":"0x4406dc086",
         "extraData":"0x476574682f4c5649562f76312e302e302f6c696e75782f676f312e342e32",
         "gasLimit":"0x1388",
         "gasUsed":"0x0",
         "hash":"0x61beeeb3e11e89d19fed2e988c8017b55c3ddb8895f531072363ce2abaf56b95",
         "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
         "miner":"0xbb7b8287f3f0a933474a79eae42cbca977791171",
         "mixHash":"0x87547a998fe63f18b36180ca918131b6b20fc5d67390e2ac2f66be3fee8fb7d2",
         "nonce":"0x1dc5b79704350bee",
         "number":"0x8b",
         "parentHash":"0x2253b8f79c23b6ff67cb2ef6fabd9ec59e1edf2d07c16d98a19378041f96624d",
         "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
         "size":"0x21f",
         "stateRoot":"0x940131b162b07452ea31b5335c4dedfdddc13338142f71f261d51dea664033b4",
         "timestamp":"0x55ba4441",
         "totalDifficulty":"0x24826e73469",
         "transactions":[
            
         ],
         "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "uncles":[
            
         ]
      }
   },
   {
      "jsonrpc":"2.0",
      "id":141,
      "result":{
         "difficulty":"0x4406dc086",
         "extraData":"0x476574682f6b6c6f737572652f76312e302e302d66633739643332642f6c696e",
         "gasLimit":"0x1388",
         "gasUsed":"0x0",
         "hash":"0xf84d9d74415364c3a7569f315ff831b910968c7dd637fffaab51278c9e7f9306",
         "logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
         "miner":"0xd7e30ae310c1d1800f5b641baa7af95b2e1fd98c",
         "mixHash":"0x6039f236ebb70ec71091df5770aef0f0faa13ef334c4c68daaffbfdf7961a3d3",
         "nonce":"0x7d8ec05d330e6e99",
         "number":"0x8b",
         "parentHash":"0x2253b8f79c23b6ff67cb2ef6fabd9ec59e1edf2d07c16d98a19378041f96624d",
         "receiptsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
         "size":"0x221",
         "stateRoot":"0x302bb7708752013f46f009dec61cad586c35dc185d20cdde0071b7487f7c2008",
         "timestamp":"0x55ba4440",
         "totalDifficulty":"0x24826e73469",
         "transactions":[
            
         ],
         "transactionsRoot":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
         "uncles":[
            
         ]
      }
   }
]`
)
