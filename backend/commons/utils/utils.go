package utils

import (
	"bytes"
	"encoding/hex"
	"log"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gobitfly/beaconchain/commons/types"

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

func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err)
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
