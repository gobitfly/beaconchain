package utils

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func mustParseUint(str string) uint64 {

	if str == "" {
		return 0
	}

	nbr, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logrus.Fatalf("fatal error parsing uint %s: %v", str, err)
	}

	return nbr
}

// // MustParseHex will parse a string into hex
// func MustParseHex(hexString string) []byte {
// 	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return data
// }

// func CORSMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Headers", "*, Authorization")
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "*")
// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

// func IsApiRequest(r *http.Request) bool {
// 	query, ok := r.URL.Query()["format"]
// 	return ok && len(query) > 0 && query[0] == "json"
// }

// var eth1AddressRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{40}$")
// var withdrawalCredentialsRE = regexp.MustCompile("^(0x)?00[0-9a-fA-F]{62}$")
// var withdrawalCredentialsAddressRE = regexp.MustCompile("^(0x)?010000000000000000000000[0-9a-fA-F]{40}$")
// var eth1TxRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{64}$")
// var zeroHashRE = regexp.MustCompile("^(0x)?0+$")
// var hashRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{96}$")

// // IsValidEth1Address verifies whether a string represents a valid eth1-address.
// func IsValidEth1Address(s string) bool {
// 	return !zeroHashRE.MatchString(s) && eth1AddressRE.MatchString(s)
// }

// // IsEth1Address verifies whether a string represents an eth1-address.
// // In contrast to IsValidEth1Address, this also returns true for the 0x0 address
// func IsEth1Address(s string) bool {
// 	return eth1AddressRE.MatchString(s)
// }

// // IsValidEth1Tx verifies whether a string represents a valid eth1-tx-hash.
// func IsValidEth1Tx(s string) bool {
// 	return !zeroHashRE.MatchString(s) && eth1TxRE.MatchString(s)
// }

// // IsEth1Tx verifies whether a string represents an eth1-tx-hash.
// // In contrast to IsValidEth1Tx, this also returns true for the 0x0 address
// func IsEth1Tx(s string) bool {
// 	return eth1TxRE.MatchString(s)
// }

// // IsHash verifies whether a string represents an eth1-hash.
// func IsHash(s string) bool {
// 	return hashRE.MatchString(s)
// }

// // IsValidWithdrawalCredentials verifies whether a string represents valid withdrawal credentials.
// func IsValidWithdrawalCredentials(s string) bool {
// 	return withdrawalCredentialsRE.MatchString(s) || withdrawalCredentialsAddressRE.MatchString(s)
// }

// // https://github.com/badoux/checkmail/blob/f9f80cb795fa/checkmail.go#L37
// var emailRE = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// // IsValidEmail verifies whether a string represents a valid email-address.
// func IsValidEmail(s string) bool {
// 	return emailRE.MatchString(s)
// }

// // IsValidUrl verifies whether a string represents a valid Url.
// func IsValidUrl(s string) bool {
// 	u, err := url.ParseRequestURI(s)
// 	if err != nil {
// 		return false
// 	}
// 	if u.Scheme != "http" && u.Scheme != "https" {
// 		return false
// 	}
// 	if len(u.Host) == 0 {
// 		return false
// 	}
// 	return govalidator.IsURL(s)
// }

// // RoundDecimals rounds (nearest) a number to the specified number of digits after comma
// func RoundDecimals(f float64, n int) float64 {
// 	d := math.Pow10(n)
// 	return math.Round(f*d) / d
// }

// // HashAndEncode digests the input with sha256 and returns it as hex string
// func HashAndEncode(input string) string {
// 	codeHashedBytes := sha256.Sum256([]byte(input))
// 	return hex.EncodeToString(codeHashedBytes[:])
// }

// const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// // RandomString returns a random hex-string
// func RandomString(length int) string {
// 	b, _ := GenerateRandomBytesSecure(length)
// 	for i := range b {
// 		b[i] = charset[int(b[i])%len(charset)]
// 	}
// 	return string(b)
// }

// func GenerateRandomBytesSecure(n int) ([]byte, error) {
// 	b := make([]byte, n)
// 	_, err := securerand.Read(b)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return b, nil
// }

// func SqlRowsToJSON(rows *sql.Rows) ([]interface{}, error) {
// 	columnTypes, err := rows.ColumnTypes()

// 	if err != nil {
// 		return nil, fmt.Errorf("error getting column types: %w", err)
// 	}

// 	count := len(columnTypes)
// 	finalRows := []interface{}{}

// 	for rows.Next() {

// 		scanArgs := make([]interface{}, count)

// 		for i, v := range columnTypes {
// 			switch v.DatabaseTypeName() {
// 			case "VARCHAR", "TEXT", "UUID":
// 				scanArgs[i] = new(sql.NullString)
// 			case "BOOL":
// 				scanArgs[i] = new(sql.NullBool)
// 			case "INT4", "INT8":
// 				scanArgs[i] = new(sql.NullInt64)
// 			case "FLOAT8":
// 				scanArgs[i] = new(sql.NullFloat64)
// 			case "TIMESTAMP":
// 				scanArgs[i] = new(sql.NullTime)
// 			case "_INT4", "_INT8":
// 				scanArgs[i] = new(pq.Int64Array)
// 			default:
// 				scanArgs[i] = new(sql.NullString)
// 			}
// 		}

// 		err := rows.Scan(scanArgs...)

// 		if err != nil {
// 			return nil, fmt.Errorf("error scanning rows: %w", err)
// 		}

// 		masterData := map[string]interface{}{}

// 		for i, v := range columnTypes {

// 			//log.Println(v.Name(), v.DatabaseTypeName())
// 			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
// 				if z.Valid {
// 					masterData[v.Name()] = z.Bool
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
// 				if z.Valid {
// 					if v.DatabaseTypeName() == "BYTEA" {
// 						if len(z.String) > 0 {
// 							masterData[v.Name()] = "0x" + hex.EncodeToString([]byte(z.String))
// 						} else {
// 							masterData[v.Name()] = nil
// 						}
// 					} else if v.DatabaseTypeName() == "NUMERIC" {
// 						nbr, _ := new(big.Int).SetString(z.String, 10)
// 						masterData[v.Name()] = nbr
// 					} else {
// 						masterData[v.Name()] = z.String
// 					}
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
// 				if z.Valid {
// 					masterData[v.Name()] = z.Int64
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
// 				if z.Valid {
// 					masterData[v.Name()] = z.Int32
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
// 				if z.Valid {
// 					masterData[v.Name()] = z.Float64
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			if z, ok := (scanArgs[i]).(*sql.NullTime); ok {
// 				if z.Valid {
// 					masterData[v.Name()] = z.Time.Unix()
// 				} else {
// 					masterData[v.Name()] = nil
// 				}
// 				continue
// 			}

// 			masterData[v.Name()] = scanArgs[i]
// 		}

// 		finalRows = append(finalRows, masterData)
// 	}

// 	return finalRows, nil
// }

// func GetSigningDomain() ([]byte, error) {
// 	beaconConfig := prysm_params.BeaconConfig()
// 	genForkVersion, err := hex.DecodeString(strings.Replace(Config.Chain.ClConfig.GenesisForkVersion, "0x", "", -1))
// 	if err != nil {
// 		return nil, err
// 	}

// 	domain, err := signing.ComputeDomain(
// 		beaconConfig.DomainDeposit,
// 		genForkVersion,
// 		beaconConfig.ZeroHash[:],
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return domain, err
// }

func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.ClConfig.SecondsPerSlot), 0)
}

func BitAtVector(b []byte, i int) bool {
	bb := b[i/8]
	return (bb & (1 << uint(i%8))) > 0
}

func BitAtVectorReversed(b []byte, i int) bool {
	bb := b[i/8]
	return (bb & (1 << uint(7-(i%8)))) > 0
}

func GetNetwork() string {
	return strings.ToLower(Config.Chain.ClConfig.ConfigName)
}

func ElementExists(arr []string, el string) bool {
	for _, e := range arr {
		if e == el {
			return true
		}
	}
	return false
}
