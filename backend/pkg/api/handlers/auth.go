package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const (
	authenticatedKey = "authenticated"
	userIdKey        = "user_id"
	subscriptionKey  = "subscription"
	userGroupKey     = "user_group"
)

func (h *HandlerService) getUser(r *http.Request) (types.User, error) {
	authenticated := h.scs.GetBool(r.Context(), authenticatedKey)
	if !authenticated {
		return types.User{}, newUnauthorizedErr("not authenticated")
	}
	subscription := h.scs.GetString(r.Context(), subscriptionKey)
	userGroup := h.scs.GetString(r.Context(), userGroupKey)
	userId, ok := h.scs.Get(r.Context(), userIdKey).(uint64)
	if !ok {
		return types.User{}, errors.New("error parsind user id from session, not a uint64")
	}

	return types.User{
		Id:        userId,
		ProductId: subscription,
		UserGroup: userGroup,
	}, nil
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

func (h *HandlerService) InternalPostLogin(w http.ResponseWriter, r *http.Request) {
	// validate request
	var v validationError
	req := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if bodyErr := v.checkBody(&req, r.Body); bodyErr != nil {
		handleErr(w, bodyErr)
		return
	}

	email := v.checkEmail(req.Email)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	badCredentialsErr := newUnauthorizedErr("invalid email or password")
	// fetch user
	user, err := h.dai.GetUserInfo(email)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			err = badCredentialsErr
		}
		handleErr(w, err)
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

// Middlewares

// checks if user has access to dashboard
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

		user, err := h.getUser(r)
		if err != nil {
			handleErr(w, err)
			return
		}
		dashboard, err := h.dai.GetValidatorDashboardInfo(types.VDBIdPrimary(dashboardId))
		if err != nil {
			handleErr(w, err)
			return
		}

		if dashboard.UserId != user.Id {
			// user does not have access to dashboard, return 404 to avoid leaking information
			returnNotFound(w, fmt.Errorf("%w: dashboard with id %v not found", dataaccess.ErrNotFound, dashboardId))
			return
		}
		next.ServeHTTP(w, r)
	})
}
