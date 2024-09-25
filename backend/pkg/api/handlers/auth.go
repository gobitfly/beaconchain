package handlers

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/mail"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	commonTypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/userservice"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const (
	authenticatedKey = "authenticated"
	userIdKey        = "user_id"
	subscriptionKey  = "subscription"
	userGroupKey     = "user_group"
	mobileAuthKey    = "mobile_auth"
)

const authConfirmEmailRateLimit = time.Minute * 2
const authResetEmailRateLimit = time.Minute * 2
const authEmailExpireTime = time.Minute * 30

type ctxKet string

const ctxUserIdKey ctxKet = "user_id"

var errBadCredentials = newUnauthorizedErr("invalid email or password")

func (h *HandlerService) getUserBySession(r *http.Request) (types.UserCredentialInfo, error) {
	authenticated := h.scs.GetBool(r.Context(), authenticatedKey)
	if !authenticated {
		return types.UserCredentialInfo{}, newUnauthorizedErr("not authenticated")
	}
	subscription := h.scs.GetString(r.Context(), subscriptionKey)
	userGroup := h.scs.GetString(r.Context(), userGroupKey)
	userId, ok := h.scs.Get(r.Context(), userIdKey).(uint64)
	if !ok {
		return types.UserCredentialInfo{}, errors.New("error parsing user id from session, not a uint64")
	}

	return types.UserCredentialInfo{
		Id:        userId,
		ProductId: subscription,
		UserGroup: userGroup,
	}, nil
}

func (h *HandlerService) purgeAllSessionsForUser(ctx context.Context, userId uint64) error {
	// invalidate all sessions for this user
	err := h.scs.Iterate(ctx, func(ctx context.Context) error {
		sessionUserID, ok := h.scs.Get(ctx, userIdKey).(uint64)
		if !ok {
			log.Error(nil, "error parsing user id from session, not a uint64", 0, nil)
			return nil
		}

		if userId == sessionUserID {
			return h.scs.Destroy(ctx)
		}

		return nil
	})

	return err
}

// TODO move to service?
func (h *HandlerService) sendConfirmationEmail(ctx context.Context, userId uint64, email string) error {
	// 1. check last confirmation time to enforce ratelimit
	lastTs, err := h.dai.GetEmailConfirmationTime(ctx, userId)
	if err != nil {
		return errors.New("error getting confirmation-ts")
	}
	if lastTs.Add(authConfirmEmailRateLimit).After(time.Now()) {
		return newTooManyRequestsErr("rate limit reached, try again later")
	}

	// 2. update confirmation hash (before sending so there's no hash mismatch on failure)
	confirmationHash := utils.RandomString(40)
	err = h.dai.UpdateEmailConfirmationHash(ctx, userId, email, confirmationHash)
	if err != nil {
		return errors.New("error updating confirmation hash")
	}

	// 3. send confirmation email
	subject := fmt.Sprintf("%s: Verify your email-address", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`Please verify your email on %[1]s by clicking this link:

https://%[1]s/api/i/users/email-confirmations/%[2]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, confirmationHash)
	err = mail.SendTextMail(email, subject, msg, []commonTypes.EmailAttachment{})
	if err != nil {
		return errors.New("error sending confirmation email, try again later")
	}

	// 4. update confirmation time (only after mail was sent)
	err = h.dai.UpdateEmailConfirmationTime(ctx, userId)
	if err != nil {
		// shouldn't present this as error to user, confirmation works fine
		log.Error(err, "error updating email confirmation time, rate limiting won't be enforced", 0, nil)
	}
	return nil
}

// TODO move to service?
func (h *HandlerService) sendPasswordResetEmail(ctx context.Context, userId uint64, email string) error {
	// 0. check if password resets are allowed
	// (can be forbidden by admin (not yet in v2))
	passwordResetAllowed, err := h.dai.IsPasswordResetAllowed(ctx, userId)
	if err != nil {
		return err
	}
	if !passwordResetAllowed {
		return newForbiddenErr("password reset not allowed")
	}

	// 1. check last confirmation time to enforce ratelimit
	lastTs, err := h.dai.GetPasswordResetTime(ctx, userId)
	if err != nil {
		return errors.New("error getting confirmation-ts")
	}
	if lastTs.Add(authResetEmailRateLimit).After(time.Now()) {
		return newTooManyRequestsErr("rate limit reached, try again later")
	}

	// 2. update reset hash (before sending so there's no hash mismatch on failure)
	resetHash := utils.RandomString(40)
	err = h.dai.UpdatePasswordResetHash(ctx, userId, resetHash)
	if err != nil {
		return errors.New("error updating confirmation hash")
	}

	// 3. send confirmation email
	subject := fmt.Sprintf("%s: Reset your passsword", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`Please reset your password on %[1]s by clicking this link:

https://%[1]s/reset-password/%[2]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, resetHash)
	err = mail.SendTextMail(email, subject, msg, []commonTypes.EmailAttachment{})
	if err != nil {
		return errors.New("error sending reset email, try again later")
	}

	// 4. update reset time (only after mail was sent)
	err = h.dai.UpdatePasswordResetTime(ctx, userId)
	if err != nil {
		// shouldn't present this as error to user, reset works fine
		log.Error(err, "error updating password reset time, rate limiting won't be enforced", 0, nil)
	}
	return nil
}

func (h *HandlerService) GetUserIdBySession(r *http.Request) (uint64, error) {
	user, err := h.getUserBySession(r)
	return user.Id, err
}

const authHeaderPrefix = "Bearer "

func (h *HandlerService) GetUserIdByApiKey(r *http.Request) (uint64, error) {
	// TODO: store user id in context during ratelimting and use it here
	var apiKey string
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, authHeaderPrefix) {
		apiKey = strings.TrimPrefix(authHeader, authHeaderPrefix)
	} else {
		apiKey = r.URL.Query().Get("api_key")
	}
	if apiKey == "" {
		return 0, newUnauthorizedErr("missing api key")
	}
	userId, err := h.dai.GetUserIdByApiKey(r.Context(), apiKey)
	if errors.Is(err, dataaccess.ErrNotFound) {
		err = newUnauthorizedErr("api key not found")
	}
	return userId, err
}

// if this is used, user ID should've been stored in context (by GetUserIdStoreMiddleware)
func GetUserIdByContext(r *http.Request) (uint64, error) {
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		return 0, newUnauthorizedErr("user not authenticated")
	}
	return userId, nil
}

// Handlers

func (h *HandlerService) InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostUsers(w http.ResponseWriter, r *http.Request) {
	// validate request
	var v validationError
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}

	// validate email
	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	_, err := h.dai.GetUserByEmail(r.Context(), email)
	if !errors.Is(err, dataaccess.ErrNotFound) {
		if err == nil {
			returnConflict(w, r, errors.New("email already registered"))
		} else {
			handleErr(w, r, err)
		}
		return
	}

	// validate password
	password := v.checkPassword(req.Password)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		handleErr(w, r, errors.New("error hashing password"))
		return
	}

	// add user
	userId, err := h.dai.CreateUser(r.Context(), email, string(passwordHash))
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// email confirmation
	err = h.sendConfirmationEmail(r.Context(), userId, email)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnOk(w, r, nil)
}

// email confirmations + changes
func (h *HandlerService) InternalPostUserConfirm(w http.ResponseWriter, r *http.Request) {
	var v validationError
	confirmationHash := v.checkUserEmailToken(mux.Vars(r)["token"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	userId, err := h.dai.GetUserIdByConfirmationHash(r.Context(), confirmationHash)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	confirmationTime, err := h.dai.GetEmailConfirmationTime(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if confirmationTime.Add(authEmailExpireTime).Before(time.Now()) {
		handleErr(w, r, newGoneErr("confirmation link expired"))
		return
	}

	err = h.dai.UpdateUserEmail(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	err = h.purgeAllSessionsForUser(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) InternalPostUserPasswordReset(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Email string `json:"email"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}

	// validate email
	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	userId, err := h.dai.GetUserByEmail(r.Context(), email)
	if err != nil {
		if err == dataaccess.ErrNotFound {
			// don't leak if email is registered
			returnOk(w, r, nil)
		} else {
			handleErr(w, r, err)
		}
		return
	}

	// send password reset email
	err = h.sendPasswordResetEmail(r.Context(), userId, email)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostUserPasswordResetHash(w http.ResponseWriter, r *http.Request) {
	var v validationError
	resetToken := v.checkUserEmailToken(mux.Vars(r)["token"])
	req := struct {
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	password := v.checkPassword(req.Password)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	// check token validity
	userId, err := h.dai.GetUserIdByResetHash(r.Context(), resetToken)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	resetTime, err := h.dai.GetPasswordResetTime(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if resetTime.Add(authEmailExpireTime).Before(time.Now()) {
		handleErr(w, r, newGoneErr("reset link expired"))
		return
	}

	// change password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		handleErr(w, r, errors.New("error hashing password"))
		return
	}
	err = h.dai.UpdateUserPassword(r.Context(), userId, string(passwordHash))
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// if email is not confirmed, confirm since they clicked a link emailed to them
	userInfo, err := h.dai.GetUserCredentialInfo(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !userInfo.EmailConfirmed {
		err = h.dai.UpdateUserEmail(r.Context(), userId)
		if err != nil {
			handleErr(w, r, err)
			return
		}
	}

	err = h.purgeAllSessionsForUser(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) InternalPostLogin(w http.ResponseWriter, r *http.Request) {
	// validate request
	var v validationError
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}

	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	// fetch user
	userId, err := h.dai.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = errBadCredentials
		}
		handleErr(w, r, err)
		return
	}
	user, err := h.dai.GetUserCredentialInfo(r.Context(), userId)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = errBadCredentials
		}
		handleErr(w, r, err)
		return
	}
	if !user.EmailConfirmed {
		handleErr(w, r, newUnauthorizedErr("email not confirmed"))
		return
	}

	// validate password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		handleErr(w, r, errBadCredentials)
		return
	}

	// change privileges
	err = h.scs.RenewToken(r.Context())
	if err != nil {
		handleErr(w, r, errors.New("error creating session"))
		return
	}

	h.scs.Put(r.Context(), authenticatedKey, true)
	h.scs.Put(r.Context(), userIdKey, user.Id)
	h.scs.Put(r.Context(), subscriptionKey, user.ProductId)
	h.scs.Put(r.Context(), userGroupKey, user.UserGroup)

	returnOk(w, r, nil)
}

// Can be used to login on mobile, requires an authenticated session
// Response must conform to OAuth spec
func (h *HandlerService) InternalPostMobileAuthorize(w http.ResponseWriter, r *http.Request) {
	req := struct {
		DeviceIDAndName string `json:"client_id"`
		RedirectURI     string `json:"redirect_uri"`
		State           string `json:"state"`
	}{}

	// Retrieve parameters from GET request
	req.DeviceIDAndName = r.URL.Query().Get("client_id")
	req.RedirectURI = r.URL.Query().Get("redirect_uri")
	req.State = r.URL.Query().Get("state")

	// To be compliant with OAuth 2 Spec, we include client_name in client_id instead of adding an additional param
	// Split req.DeviceID on ":", first one is the client id and second one the client name
	deviceIDParts := strings.Split(req.DeviceIDAndName, ":")
	var clientID, clientName string
	if len(deviceIDParts) != 2 {
		clientID = req.DeviceIDAndName
		clientName = "Unknown"
	} else {
		clientID = deviceIDParts[0]
		clientName = deviceIDParts[1]
	}

	state := ""
	if req.State != "" {
		state = "&state=" + req.State
	}

	// check if user has a session
	userInfo, err := h.getUserBySession(r)
	if err != nil {
		callback := req.RedirectURI + "?error=invalid_request&error_description=unauthorized_client" + state
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}

	// check if oauth app exists to validate whether redirect uri is valid
	appInfo, err := h.dai.GetAppDataFromRedirectUri(req.RedirectURI)
	if err != nil {
		callback := req.RedirectURI + "?error=invalid_request&error_description=missing_redirect_uri" + state
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}

	// renew session and pass to callback
	err = h.scs.RenewToken(r.Context())
	if err != nil {
		callback := req.RedirectURI + "?error=invalid_request&error_description=server_error" + state
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}
	session := h.scs.Token(r.Context())

	sanitizedDeviceName := html.EscapeString(clientName)
	err = h.dai.AddUserDevice(userInfo.Id, utils.HashAndEncode(session+session), clientID, sanitizedDeviceName, appInfo.ID)
	if err != nil {
		log.Warnf("Error adding user device: %v", err)
		callback := req.RedirectURI + "?error=invalid_request&error_description=server_error" + state
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}

	// pass via redirect to app oauth callback handler
	callback := req.RedirectURI + "?access_token=" + session + "&token_type=bearer" + state // prefixed session
	http.Redirect(w, r, callback, http.StatusFound)
}

// Abstract: One time Transitions old v1 app sessions to new v2 sessions so users stay signed in
// Can be used to exchange a legacy mobile auth access_token & refresh_token pair for a session
// Refresh token is consumed and can no longer be used after this
func (h *HandlerService) InternalPostMobileEquivalentExchange(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		DeviceName   string `json:"client_name"`
		RefreshToken string `json:"refresh_token"`
		DeviceID     string `json:"client_id"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	// get user id by refresh token
	userID, refreshTokenHashed, err := h.getTokenByRefresh(r, req.RefreshToken)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// Get user info
	user, err := h.dai.GetUserCredentialInfo(r.Context(), userID)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = errBadCredentials
		}
		handleErr(w, r, err)
		return
	}
	if !user.EmailConfirmed {
		handleErr(w, r, newUnauthorizedErr("email not confirmed"))
		return
	}

	// create new session
	err = h.scs.RenewToken(r.Context())
	if err != nil {
		handleErr(w, r, errors.New("error creating session"))
		return
	}
	session := h.scs.Token(r.Context())

	// invalidate old refresh token and replace with hashed session id
	sanitizedDeviceName := html.EscapeString(req.DeviceName)
	err = h.dai.MigrateMobileSession(refreshTokenHashed, utils.HashAndEncode(session+session), req.DeviceID, sanitizedDeviceName) // salted with session
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// set fields of session after invalidating refresh token
	h.scs.Put(r.Context(), authenticatedKey, true)
	h.scs.Put(r.Context(), userIdKey, userID)
	h.scs.Put(r.Context(), subscriptionKey, user.ProductId)
	h.scs.Put(r.Context(), userGroupKey, user.UserGroup)
	h.scs.Put(r.Context(), mobileAuthKey, true)

	returnOk(w, r, struct {
		Session string
	}{
		Session: session,
	})
}

func (h *HandlerService) InternalPostUsersMeNotificationSettingsPairedDevicesToken(w http.ResponseWriter, r *http.Request) {
	deviceID := mux.Vars(r)["client_id"]
	var v validationError
	req := struct {
		Token string `json:"token"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	err = h.dai.AddMobileNotificationToken(user.Id, deviceID, req.Token)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnOk(w, r, nil)
}

const USER_SUBSCRIPTION_LIMIT = 8

func (h *HandlerService) InternalHandleMobilePurchase(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := types.MobileSubscription{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	if req.ProductIDUnverified == "plankton" {
		handleErr(w, r, newForbiddenErr("plankton subscription has been discontinued"))
		return
	}

	// Only allow ios and android purchases to be registered via this endpoint
	if req.Transaction.Type != "ios-appstore" && req.Transaction.Type != "android-playstore" {
		handleErr(w, r, newForbiddenErr("only ios-appstore and android-playstore purchases are allowed"))
		return
	}

	subscriptionCount, err := h.dai.GetAppSubscriptionCount(user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if subscriptionCount >= USER_SUBSCRIPTION_LIMIT {
		handleErr(w, r, newForbiddenErr("user has reached the subscription limit"))
		return
	}

	// Verify subscription with apple/google
	verifyPackage := &commonTypes.PremiumData{
		ID:        0,
		Receipt:   req.Transaction.Receipt,
		Store:     req.Transaction.Type,
		Active:    false,
		ProductID: req.ProductIDUnverified,
		ExpiresAt: time.Now(),
	}

	validationResult, err := userservice.VerifyReceipt(nil, nil, verifyPackage)
	if err != nil {
		log.Warn(err, "could not verify receipt %v", 0, map[string]interface{}{"receipt": verifyPackage.Receipt})
		metrics.Errors.WithLabelValues(fmt.Sprintf("appsub_verify_%s_failed", req.Transaction.Type)).Inc()
		if errors.Is(err, userservice.ErrClientInit) {
			log.Error(err, "Apple or Google client is NOT initialized. Did you provide their configuration?", 0, nil)
			handleErr(w, r, err)
			return
		}
	}

	err = h.dai.AddMobilePurchase(nil, user.Id, req, validationResult, "")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	if !validationResult.Valid {
		handleErr(w, r, newForbiddenErr("receipt is not valid"))
		return
	}

	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostLogout(w http.ResponseWriter, r *http.Request) {
	err := h.scs.Destroy(r.Context())
	if err != nil {
		handleErr(w, r, err)
		return
	}
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalDeleteUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// TODO allow if user has any subsciptions etc?
	err = h.dai.RemoveUser(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	err = h.purgeAllSessionsForUser(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) InternalPostUserEmail(w http.ResponseWriter, r *http.Request) {
	// validate user
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	userInfo, err := h.dai.GetUserCredentialInfo(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !userInfo.EmailConfirmed {
		handleErr(w, r, newConflictErr("email not confirmed"))
		return
	}

	// validate request
	var v validationError
	req := struct {
		Email    string `json:"new_email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}

	// validate new email
	newEmail := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	if newEmail == userInfo.Email {
		handleErr(w, r, newConflictErr("can't reuse current email"))
		return
	}

	_, err = h.dai.GetUserByEmail(r.Context(), newEmail)
	if !errors.Is(err, dataaccess.ErrNotFound) {
		if err == nil {
			handleErr(w, r, newConflictErr("email already registered"))
		} else {
			handleErr(w, r, err)
		}
		return
	}

	// validate password
	password := v.checkPassword(req.Password)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		handleErr(w, r, newUnauthorizedErr("invalid password"))
		return
	}

	// email confirmation
	err = h.sendConfirmationEmail(r.Context(), userInfo.Id, newEmail)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalPostUserEmailResponse{
		Data: types.EmailUpdate{
			Id:           userInfo.Id,
			CurrentEmail: userInfo.Email,
			PendingEmail: newEmail,
		},
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutUserPassword(w http.ResponseWriter, r *http.Request) {
	// validate user
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	// user doesn't contain password, fetch from db
	userData, err := h.dai.GetUserCredentialInfo(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	// validate request
	var v validationError
	req := struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}

	// validate passwords
	oldPassword := v.checkPassword(req.OldPassword)
	newPassword := v.checkPassword(req.NewPassword)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(oldPassword))
	if err != nil {
		handleErr(w, r, errors.New("invalid password"))
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		handleErr(w, r, errors.New("error hashing password"))
		return
	}

	// change password
	err = h.dai.UpdateUserPassword(r.Context(), user.Id, string(passwordHash))
	if err != nil {
		handleErr(w, r, err)
		return
	}

	err = h.purgeAllSessionsForUser(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

// Middlewares

// returns a middleware that stores user id in context, using the provided function
func GetUserIdStoreMiddleware(userIdFunc func(r *http.Request) (uint64, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userId, err := userIdFunc(r)
			if err != nil {
				if errors.Is(err, errUnauthorized) {
					// if next handler requires authentication, it should return 'unauthorized' itself
					next.ServeHTTP(w, r)
				} else {
					handleErr(w, r, err)
				}
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserIdKey, userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// returns a middleware that checks if user has access to dashboard when a primary id is used
func (h *HandlerService) VDBAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		dashboardId, err := strconv.ParseUint(mux.Vars(r)["dashboard_id"], 10, 64)
		if err != nil {
			// if primary id is not used, no need to check access
			next.ServeHTTP(w, r)
			return
		}
		// primary id is used -> user needs to have access to dashboard

		userId, err := GetUserIdByContext(r)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		// store user id in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserIdKey, userId)
		r = r.WithContext(ctx)

		dashboardUser, err := h.dai.GetValidatorDashboardUser(r.Context(), types.VDBIdPrimary(dashboardId))
		if err != nil {
			handleErr(w, r, err)
			return
		}

		if dashboardUser.UserId != userId {
			// user does not have access to dashboard
			// the proper error would be 403 Forbidden, but we don't want to leak information so we return 404 Not Found
			handleErr(w, r, newNotFoundErr("dashboard with id %v not found", dashboardId))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// returns a middleware that checks if user has premium perk to use public validator dashboard api
// in the middleware chain, this should be used after GetVDBAuthMiddleware
func (h *HandlerService) ManageViaApiCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get user id from context
		userId, err := GetUserIdByContext(r)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		if !userInfo.PremiumPerks.ManageDashboardViaApi {
			handleErr(w, r, newForbiddenErr("user does not have access to public validator dashboard endpoints"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// middleware check to return if specified dashboard is not archived (and accessible)
func (h *HandlerService) VDBArchivedCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
		if err != nil {
			handleErr(w, r, err)
			return
		}
		if len(dashboardId.Validators) > 0 {
			next.ServeHTTP(w, r)
			return
		}
		dashboard, err := h.dai.GetValidatorDashboardInfo(r.Context(), dashboardId.Id)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		if dashboard.IsArchived {
			handleErr(w, r, newForbiddenErr("dashboard with id %v is archived", dashboardId))
			return
		}
		next.ServeHTTP(w, r)
	})
}
