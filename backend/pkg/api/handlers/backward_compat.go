package handlers

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/golang-jwt/jwt"
)

var signingMethod = jwt.SigningMethodHS256

type CustomClaims struct {
	UserID   uint64 `json:"userID"`
	AppID    uint64 `json:"appID"`
	DeviceID uint64 `json:"deviceID"`
	Package  string `json:"package"`
	Theme    string `json:"theme"`
	jwt.StandardClaims
}

func (h *HandlerService) getTokenByRefresh(r *http.Request, refreshToken string) (uint64, string, error) {
	accessToken := r.Header.Get("Authorization")

	// hash refreshtoken
	refreshTokenHashed := utils.HashAndEncode(refreshToken)

	// Extract userId from JWT. Note that this is just an unvalidated claim!
	// Do not use userIDClaim as userID until confirmed by refreshToken validation
	unsafeClaims, err := UnsafeGetClaims(accessToken)
	if err != nil {
		log.Warnf("Error getting claims from access token: %v", err)
		return 0, "", newUnauthorizedErr("invalid token")
	}

	log.Infof("refresh token: %v, claims: %v, hashed refresh: %v", refreshToken, unsafeClaims, refreshTokenHashed)

	// confirm all claims via db lookup and refreshtoken check
	userID, err := h.dai.GetUserIdByRefreshToken(unsafeClaims.UserID, unsafeClaims.AppID, unsafeClaims.DeviceID, refreshTokenHashed)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", dataaccess.ErrNotFound
		}
		return 0, "", errors.Wrap(err, "Error getting user by refresh token")
	}

	return userID, refreshTokenHashed, nil
}

// UnsafeGetClaims this method returns the userID of a given jwt token WITHOUT VALIDATION
// DO NOT USE THIS METHOD AS RELIABLE SOURCE FOR USERID
func UnsafeGetClaims(tokenString string) (*CustomClaims, error) {
	return accessTokenGetClaims(tokenString, false)
}

func stripOffBearerFromToken(tokenString string) string {
	if len(tokenString) > 6 && strings.ToUpper(tokenString[0:6]) == "BEARER" {
		return tokenString[7:]
	}
	return tokenString //"", errors.New("Only bearer tokens are supported, got: " + tokenString)
}

func accessTokenGetClaims(tokenStringFull string, validate bool) (*CustomClaims, error) {
	tokenString := stripOffBearerFromToken(tokenStringFull)

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getSignKey()
	})

	if err != nil && validate {
		if !strings.Contains(err.Error(), "token is expired") && token != nil {
			log.Warnf("Error parsing token: %v", err)
		}

		return nil, err
	}

	if token == nil {
		return nil, fmt.Errorf("error token is not defined %v", tokenStringFull)
	}

	// Make sure header hasnt been tampered with
	if token.Method != signingMethod {
		return nil, errors.New("only SHA256hmac as signature method is allowed")
	}

	claims, ok := token.Claims.(*CustomClaims)

	// Check issuer claim
	if claims.Issuer != utils.Config.Frontend.JwtIssuer {
		return nil, errors.New("invalid issuer claim")
	}

	valid := ok && token.Valid

	if valid || !validate {
		return claims, nil
	}

	return nil, errors.New("token validity or claims cannot be verified")
}

func getSignKey() ([]byte, error) {
	signSecret, err := hex.DecodeString(utils.Config.Frontend.JwtSigningSecret)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding jwtSecretKey, not in hex format or missing from config?")
	}
	return signSecret, nil
}
