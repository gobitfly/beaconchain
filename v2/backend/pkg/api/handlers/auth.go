package handlers

import (
	"context"
	"database/sql"
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
	commontsTypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
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
const authEmailExpireTime = time.Minute * 30

type ctxKet string

const ctxUserIdKey ctxKet = "user_id"

func (h *HandlerService) getUserBySession(r *http.Request) (types.UserCredentialInfo, error) {
	authenticated := h.scs.GetBool(r.Context(), authenticatedKey)
	if !authenticated {
		return types.UserCredentialInfo{}, newUnauthorizedErr("not authenticated")
	}
	subscription := h.scs.GetString(r.Context(), subscriptionKey)
	userGroup := h.scs.GetString(r.Context(), userGroupKey)
	userId, ok := h.scs.Get(r.Context(), userIdKey).(uint64)
	if !ok {
		return types.UserCredentialInfo{}, errors.New("error parsind user id from session, not a uint64")
	}

	return types.UserCredentialInfo{
		Id:        userId,
		ProductId: subscription,
		UserGroup: userGroup,
	}, nil
}

// TODO move to service?
func (h *HandlerService) sendConfirmationEmail(ctx context.Context, userId uint64, email string) error {
	// 1. check last confirmation time to enforce ratelimit
	lastTs, err := h.dai.GetEmailConfirmationTime(ctx, userId)
	if err != nil {
		return errors.New("error getting confirmation-ts")
	}
	if lastTs.Add(authConfirmEmailRateLimit).After(time.Now()) {
		return errors.New("rate limit reached, try again later")
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
	err = mail.SendTextMail(email, subject, msg, []commontsTypes.EmailAttachment{})
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

func (h *HandlerService) GetUserIdBySession(r *http.Request) (uint64, error) {
	user, err := h.getUserBySession(r)
	if err != nil {
		return 0, err
	}
	return user.Id, nil
}

const authHeaderPrefix = "Bearer "

func (h *HandlerService) GetUserIdByApiKey(r *http.Request) (uint64, error) {
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

// Handlers

func (h *HandlerService) InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostUsers(w http.ResponseWriter, r *http.Request) {
	// validate request
	var v validationError
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	// validate email
	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	_, err := h.dai.GetUserByEmail(r.Context(), email)
	if !errors.Is(err, dataaccess.ErrNotFound) {
		if err == nil {
			returnConflict(w, errors.New("email already registered"))
		} else {
			handleErr(w, err)
		}
		return
	}

	// validate password
	password := v.checkPassword(req.Password)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		handleErr(w, errors.New("error hashing password"))
		return
	}

	// add user
	userId, err := h.dai.CreateUser(r.Context(), email, string(passwordHash))
	if err != nil {
		handleErr(w, err)
		return
	}

	// email confirmation
	err = h.sendConfirmationEmail(r.Context(), userId, email)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnOk(w, nil)
}

// email confirmations + changes
func (h *HandlerService) InternalPostUserConfirm(w http.ResponseWriter, r *http.Request) {
	var v validationError
	confirmationHash := v.checkConfirmationHash(mux.Vars(r)["token"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	userId, err := h.dai.GetUserIdByConfirmationHash(confirmationHash)
	if err != nil {
		handleErr(w, err)
		return
	}
	confirmationTime, err := h.dai.GetEmailConfirmationTime(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if confirmationTime.Add(authEmailExpireTime).Before(time.Now()) {
		handleErr(w, errors.New("confirmation link expired"))
		return
	}

	err = h.dai.UpdateUserEmail(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}

	// TODO: purge all user sessions

	returnOk(w, nil)
}

func (h *HandlerService) InternalPostLogin(w http.ResponseWriter, r *http.Request) {
	// validate request
	var v validationError
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	badCredentialsErr := newUnauthorizedErr("invalid email or password")
	// fetch user
	userId, err := h.dai.GetUserByEmail(r.Context(), email)
	if err != nil {
		handleErr(w, err)
		return
	}
	user, err := h.dai.GetUserCredentialInfo(r.Context(), userId)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = badCredentialsErr
		}
		handleErr(w, err)
		return
	}
	if !user.EmailConfirmed {
		handleErr(w, newUnauthorizedErr("email not confirmed"))
		return
	}

	// validate password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		handleErr(w, badCredentialsErr)
		return
	}

	// change privileges
	err = h.scs.RenewToken(r.Context())
	if err != nil {
		handleErr(w, errors.New("error creating session"))
		return
	}

	h.scs.Put(r.Context(), authenticatedKey, true)
	h.scs.Put(r.Context(), userIdKey, user.Id)
	h.scs.Put(r.Context(), subscriptionKey, user.ProductId)
	h.scs.Put(r.Context(), userGroupKey, user.UserGroup)

	returnOk(w, nil)
}

// Can be used to login on mobile, requires an authenticated session
// Response must conform to OAuth spec
func (h *HandlerService) InternalPostAuthorize(w http.ResponseWriter, r *http.Request) {
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
func (h *HandlerService) InternalExchangeLegacyMobileAuth(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		DeviceName   string `json:"client_name"`
		RefreshToken string `json:"refresh_token"`
		DeviceID     string `json:"client_id"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	// get user id by refresh token
	userID, refreshTokenHashed, err := h.getTokenByRefresh(r, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidTokenClaims):
			handleErr(w, newUnauthorizedErr("invalid token"))
		case errors.Is(err, sql.ErrNoRows):
			handleErr(w, dataaccess.ErrNotFound)
		default:
			handleErr(w, err)
		}
		return
	}

	// Get user info
	badCredentialsErr := newUnauthorizedErr("invalid email or password") // same error as to not leak information
	user, err := h.dai.GetUserCredentialInfo(r.Context(), userID)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = badCredentialsErr
		}
		handleErr(w, err)
		return
	}
	if !user.EmailConfirmed {
		handleErr(w, newUnauthorizedErr("email not confirmed"))
		return
	}

	// create new session
	err = h.scs.RenewToken(r.Context())
	if err != nil {
		handleErr(w, errors.New("error creating session"))
		return
	}
	session := h.scs.Token(r.Context())

	// invalidate old refresh token and replace with hashed session id
	sanitizedDeviceName := html.EscapeString(req.DeviceName)
	err = h.dai.MigrateMobileSession(refreshTokenHashed, utils.HashAndEncode(session+session), req.DeviceID, sanitizedDeviceName) // salted with session
	if err != nil {
		handleErr(w, err)
		return
	}

	// set fields of session after invalidating refresh token
	h.scs.Put(r.Context(), authenticatedKey, true)
	h.scs.Put(r.Context(), userIdKey, userID)
	h.scs.Put(r.Context(), subscriptionKey, user.ProductId)
	h.scs.Put(r.Context(), userGroupKey, user.UserGroup)
	h.scs.Put(r.Context(), mobileAuthKey, true)

	returnOk(w, struct {
		Session string
	}{
		Session: session,
	})
}

func (h *HandlerService) InternalRegisterMobilePushToken(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Token    string `json:"token"`
		DeviceID string `json:"client_id"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}

	err = h.dai.AddMobileNotificationToken(user.Id, req.DeviceID, req.Token)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnOk(w, nil)
}

func (h *HandlerService) InternalPostLogout(w http.ResponseWriter, r *http.Request) {
	err := h.scs.Destroy(r.Context())
	if err != nil {
		handleErr(w, err)
		return
	}
	returnOk(w, nil)
}

func (h *HandlerService) InternalDeleteUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}

	// TODO allow if user has any subsciptions etc?
	err = h.dai.RemoveUser(r.Context(), user.Id)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPutUserEmail(w http.ResponseWriter, r *http.Request) {
	// validate user
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserCredentialInfo(r.Context(), user.Id)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !userInfo.EmailConfirmed {
		handleErr(w, newConflictErr("email not confirmed"))
		return
	}

	// validate request
	var v validationError
	req := struct {
		Email    string `json:"new_email"`
		Password string `json:"password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	// validate new email
	newEmail := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	if newEmail == userInfo.Email {
		handleErr(w, newConflictErr("can't reuse current email"))
		return
	}

	_, err = h.dai.GetUserByEmail(r.Context(), newEmail)
	if !errors.Is(err, dataaccess.ErrNotFound) {
		if err == nil {
			handleErr(w, newConflictErr("email already registered"))
		} else {
			handleErr(w, err)
		}
		return
	}

	// validate password
	password := v.checkPassword(req.Password)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		handleErr(w, newUnauthorizedErr("invalid password"))
		return
	}

	// email confirmation
	err = h.sendConfirmationEmail(r.Context(), userInfo.Id, newEmail)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalPutUserEmailResponse{
		Data: types.EmailUpdate{
			Id:           userInfo.Id,
			CurrentEmail: userInfo.Email,
			PendingEmail: newEmail,
		},
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserPassword(w http.ResponseWriter, r *http.Request) {
	// validate user
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}

	// validate request
	var v validationError
	req := struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}

	// validate passwords
	oldPassword := v.checkPassword(req.OldPassword)
	newPassword := v.checkPassword(req.NewPassword)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		handleErr(w, errors.New("invalid password"))
		return
	}

	// change password
	err = h.dai.UpdateUserPassword(r.Context(), user.Id, newPassword)
	if err != nil {
		handleErr(w, err)
		return
	}

	// TODO: purge all user sessions

	returnNoContent(w)
}

// Middlewares

// returns a middleware that checks if user has access to dashboard when a primary id is used
// expects a userIdFunc to return user id, probably GetUserIdBySession or GetUserIdByApiKey
func (h *HandlerService) GetVDBAuthMiddleware(userIdFunc func(r *http.Request) (uint64, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			dashboardId, err := strconv.ParseUint(mux.Vars(r)["dashboard_id"], 10, 64)
			if err != nil {
				// if primary id is not used, no need to check access
				next.ServeHTTP(w, r)
				return
			}
			// primary id is used -> user needs to have access to dashboard

			userId, err := userIdFunc(r)
			if err != nil {
				handleErr(w, err)
				return
			}
			// store user id in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserIdKey, userId)
			r = r.WithContext(ctx)

			dashboard, err := h.dai.GetValidatorDashboardInfo(r.Context(), types.VDBIdPrimary(dashboardId))
			if err != nil {
				handleErr(w, err)
				return
			}

			if dashboard.UserId != userId {
				// user does not have access to dashboard
				// the proper error would be 403 Forbidden, but we don't want to leak information so we return 404 Not Found
				handleErr(w, newNotFoundErr("dashboard with id %v not found", dashboardId))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// returns a middleware that checks if user has premium perk to use public validator dashboard api
// in the middleware chain, this should be used after GetVDBAuthMiddleware
func (h *HandlerService) VDBPublicApiCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get user id from context
		userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
		if !ok {
			handleErr(w, errors.New("error getting user id from context"))
			return
		}
		userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
		if err != nil {
			handleErr(w, err)
			return
		}
		if !userInfo.PremiumPerks.ManageDashboardViaApi {
			handleErr(w, newForbiddenErr("user does not have access to public validator dashboard endpoints"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
