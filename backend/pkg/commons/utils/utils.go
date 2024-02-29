package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func mustParseUint(str string) uint64 {
	if str == "" {
		return 0
	}

	nbr, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		log.Fatal(err, "fatal error parsing uint", 0, map[string]interface{}{"str": str})
	}

	return nbr
}

func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err, "error parsing hex string", 0, map[string]interface{}{"str": hexString})
	}
	return data
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

func EpochOfSlot(slot uint64) uint64 {
	return slot / Config.Chain.ClConfig.SlotsPerEpoch
}

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func GetParentFuncName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

// sliceContains reports whether the provided string is present in the given slice of strings.
func SliceContains(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

// ForkVersionAtEpoch returns the forkversion active a specific epoch
func ForkVersionAtEpoch(epoch uint64) *types.ForkVersion {
	if epoch >= Config.Chain.ClConfig.CappellaForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.CappellaForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.CappellaForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.BellatrixForkVersion),
		}
	}
	if epoch >= Config.Chain.ClConfig.BellatrixForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.BellatrixForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.BellatrixForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.AltairForkVersion),
		}
	}
	if epoch >= Config.Chain.ClConfig.AltairForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.AltairForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.AltairForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
		}
	}
	return &types.ForkVersion{
		Epoch:           0,
		CurrentVersion:  MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
		PreviousVersion: MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
	}
}

func AddBigInts(a, b []byte) []byte {
	return new(big.Int).Add(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b)).Bytes()
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func GraffitiToString(graffiti []byte) string {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	s = strings.Replace(s, "\u0000", "", -1) // remove 0x00 bytes as it is not supported in postgres

	if !utf8.ValidString(s) {
		return "INVALID_UTF8_STRING"
	}

	return s
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}

func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// To remove all round brackets (including its content) from a string
func RemoveRoundBracketsIncludingContent(input string) string {
	openCount := 0
	result := ""
	for {
		if len(input) == 0 {
			break
		}
		openIndex := strings.Index(input, "(")
		closeIndex := strings.Index(input, ")")
		if openIndex == -1 && closeIndex == -1 {
			if openCount == 0 {
				result += input
			}
			break
		} else if openIndex != -1 && (openIndex < closeIndex || closeIndex == -1) {
			openCount++
			if openCount == 1 {
				result += input[:openIndex]
			}
			input = input[openIndex+1:]
		} else {
			if openCount > 0 {
				openCount--
			} else if openIndex == -1 && len(result) == 0 {
				result += input[:closeIndex]
			}
			input = input[closeIndex+1:]
		}
	}
	return result
}

// HashAndEncode digests the input with sha256 and returns it as hex string
func HashAndEncode(input string) string {
	codeHashedBytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(codeHashedBytes[:])
}

func SortedUniqueUint64(arr []uint64) []uint64 {
	if len(arr) <= 1 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	result := make([]uint64, 1, len(arr))
	result[0] = arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i-1] != arr[i] {
			result = append(result, arr[i])
		}
	}

	return result
}

func GetParticipatingSyncCommitteeValidators(syncAggregateBits []byte, validators []uint64) []uint64 {
	participatingValidators := []uint64{}
	for i := 0; i < len(syncAggregateBits)*8; i++ {
		val := validators[i]
		if BitAtVector(syncAggregateBits, i) {
			participatingValidators = append(participatingValidators, val)
		}
	}
	return participatingValidators
}
