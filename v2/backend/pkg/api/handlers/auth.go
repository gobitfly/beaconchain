package handlers

import (
	"context"
	"errors"
	"fmt"
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
func (h *HandlerService) sendConfirmationEmail(userId uint64, email string) error {
	// 1. check last confirmation time to enforce ratelimit
	lastTs, err := h.dai.GetEmailConfirmationTime(userId)
	if err != nil {
		return errors.New("error getting confirmation-ts")
	}
	if lastTs.Add(authConfirmEmailRateLimit).After(time.Now()) {
		return errors.New("rate limit reached, try again later")
	}

	// 2. update confirmation hash (before sending so there's no hash mismatch on failure)
	confirmationHash := utils.RandomString(40)
	err = h.dai.UpdateEmailConfirmationHash(userId, email, confirmationHash)
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
	err = h.dai.UpdateEmailConfirmationTime(userId)
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
	userId, err := h.dai.GetUserIdByApiKey(apiKey)
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

	_, err := h.dai.GetUserByEmail(email)
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
	userId, err := h.dai.CreateUser(email, string(passwordHash))
	if err != nil {
		handleErr(w, err)
		return
	}

	// email confirmation
	err = h.sendConfirmationEmail(userId, email)
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
	confirmationTime, err := h.dai.GetEmailConfirmationTime(userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if confirmationTime.Add(authEmailExpireTime).Before(time.Now()) {
		handleErr(w, errors.New("confirmation link expired"))
		return
	}

	err = h.dai.UpdateUserEmail(userId)
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
	userId, err := h.dai.GetUserByEmail(email)
	if err != nil {
		handleErr(w, err)
		return
	}
	user, err := h.dai.GetUserCredentialInfo(userId)
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
	err = h.dai.RemoveUser(user.Id)
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
	userInfo, err := h.dai.GetUserCredentialInfo(user.Id)
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

	_, err = h.dai.GetUserByEmail(newEmail)
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
	err = h.sendConfirmationEmail(userInfo.Id, newEmail)
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
	err = h.dai.UpdateUserPassword(user.Id, newPassword)
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

			dashboard, err := h.dai.GetValidatorDashboardInfo(types.VDBIdPrimary(dashboardId))
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
		userInfo, err := h.dai.GetUserInfo(userId)
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
