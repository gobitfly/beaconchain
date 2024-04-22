module github.com/gobitfly/beaconchain

go 1.21

require (
	cloud.google.com/go/bigtable v1.21.0
	cloud.google.com/go/secretmanager v1.11.5
	firebase.google.com/go v3.13.0+incompatible
	github.com/Gurpartap/storekit-go v0.0.0-20201205024111-36b6cd5c6a21
	github.com/alexedwards/scs/redisstore v0.0.0-20240316134038-7e11d57e8885
	github.com/alexedwards/scs/v2 v2.8.0
	github.com/attestantio/go-eth2-client v0.19.10
	github.com/awa/go-iap v1.26.5
	github.com/aws/aws-sdk-go-v2 v1.25.0
	github.com/aws/aws-sdk-go-v2/credentials v1.13.43
	github.com/aws/aws-sdk-go-v2/service/s3 v1.49.0
	github.com/coocood/freecache v1.2.4
	github.com/davecgh/go-spew v1.1.1
	github.com/donovanhide/eventsource v0.0.0-20210830082556-c59027999da0
	github.com/ethereum/go-ethereum v1.13.12
	github.com/go-faker/faker/v4 v4.3.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gobitfly/eth-rewards v0.1.2-0.20230403064929-411ddc40a5f7
	github.com/gobitfly/eth.store v0.0.0-20240312111708-b43f13990280
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/protobuf v1.5.3
	github.com/gomodule/redigo v1.9.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.0
	github.com/gzuidhof/tygo v0.2.13
	github.com/hashicorp/go-version v1.6.0
	github.com/hashicorp/golang-lru v1.0.2
	github.com/invopop/jsonschema v0.12.0
	github.com/jackc/pgtype v1.14.2
	github.com/jackc/pgx-shopspring-decimal v0.0.0-20220624020537-1d36b5a1853e
	github.com/jackc/pgx/v5 v5.5.3
	github.com/jmoiron/sqlx v1.3.5
	github.com/juliangruber/go-intersect v1.1.0
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.17.2
	github.com/klauspost/pgzip v1.2.6
	github.com/lib/pq v1.10.9
	github.com/mailgun/mailgun-go/v4 v4.12.0
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.18.0
	github.com/prometheus/client_golang v1.18.0
	github.com/protolambda/zrnt v0.30.0
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/prysmaticlabs/prysm/v3 v3.2.2
	github.com/rocket-pool/rocketpool-go v1.8.2
	github.com/rocket-pool/smartnode v1.11.7
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/syndtr/goleveldb v1.0.1-0.20220614013038-64ee5596c38a
	github.com/wealdtech/go-ens/v3 v3.6.0
	github.com/wealdtech/go-eth2-types/v2 v2.8.2
	github.com/wealdtech/go-eth2-util v1.8.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.19.0
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa
	golang.org/x/sync v0.6.0
	golang.org/x/text v0.14.0
	golang.org/x/tools v0.15.0
	google.golang.org/api v0.164.0
	google.golang.org/protobuf v1.32.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.112.0 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/firestore v1.14.0 // indirect
	cloud.google.com/go/iam v1.1.5 // indirect
	cloud.google.com/go/longrunning v0.5.4 // indirect
	cloud.google.com/go/storage v1.36.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.0 // indirect
	github.com/aws/smithy-go v1.20.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cncf/udpa/go v0.0.0-20220112060539-c52dc94e7fbe // indirect
	github.com/cncf/xds/go v0.0.0-20231128003011-0fa0005c9caa // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20170627025303-887ab5e44cc3 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.7.0 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.12.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/ethereum/c-kzg-4844 v0.4.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/ferranbt/fastssz v0.1.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/glendc/go-external-ip v0.1.0 // indirect
	github.com/go-chi/chi/v5 v5.0.8 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/goccy/go-yaml v1.9.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/herumi/bls-eth-go-binary v1.31.0 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/huandu/go-clone v1.6.0 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/boxo v0.8.0 // indirect
	github.com/ipfs/go-bitfield v1.1.0 // indirect
	github.com/ipfs/go-block-format v0.1.2 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-datastore v0.6.0 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.6 // indirect
	github.com/ipfs/go-ipld-format v0.4.0 // indirect
	github.com/ipfs/go-ipld-legacy v0.1.1 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipfs/go-metrics-interface v0.0.1 // indirect
	github.com/ipld/go-codec-dagpb v1.6.0 // indirect
	github.com/ipld/go-ipld-prime v0.20.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/protolambda/zssz v0.1.5 // indirect
	github.com/prysmaticlabs/fastssz v0.0.0-20221107182844-78142813af44 // indirect
	github.com/prysmaticlabs/gohashtree v0.0.2-alpha // indirect
	github.com/r3labs/sse/v2 v2.10.0 // indirect
	github.com/rs/zerolog v1.29.1 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/wealdtech/go-bytesutil v1.2.0 // indirect
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a // indirect
	github.com/wealdtech/go-multicodec v1.4.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20230126041949-52956bd4c9aa // indirect
	github.com/whyrusleeping/chunker v0.0.0-20181014151217-fe64bd25879f // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.47.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0 // indirect
	go.opentelemetry.io/otel v1.23.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.0 // indirect
	go.opentelemetry.io/otel/trace v1.23.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/oauth2 v0.17.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gonum.org/v1/gonum v0.12.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/grpc v1.62.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

replace github.com/rocket-pool/rocketpool-go v1.8.2 => github.com/gobitfly/rocketpool-go v0.0.0-20240105082836-5bb7c83a2d08

replace github.com/prysmaticlabs/prysm/v3 => github.com/gobitfly/prysm/v3 v3.0.0-20230216184552-2f3f1e8190d5
