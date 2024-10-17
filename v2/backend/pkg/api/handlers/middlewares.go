package handlers

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
)

// Middlewares

// middleware that stores user id in context, using the provided function
func StoreUserIdMiddleware(next http.Handler, userIdFunc func(r *http.Request) (uint64, error)) http.Handler {
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

		// store user id in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserIdKey, userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// middleware that stores user id in context, using the session to get the user id
func (h *HandlerService) StoreUserIdBySessionMiddleware(next http.Handler) http.Handler {
	return StoreUserIdMiddleware(next, func(r *http.Request) (uint64, error) {
		return h.GetUserIdBySession(r)
	})
}

// middleware that stores user id in context, using the api key to get the user id
func (h *HandlerService) StoreUserIdByApiKeyMiddleware(next http.Handler) http.Handler {
	return StoreUserIdMiddleware(next, func(r *http.Request) (uint64, error) {
		return h.GetUserIdByApiKey(r)
	})
}

// middleware that checks if user has access to dashboard when a primary id is used
func (h *HandlerService) VDBAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if mock data is used, no need to check access
		if isMockEnabled, ok := r.Context().Value(ctxIsMockedKey).(bool); ok && isMockEnabled {
			next.ServeHTTP(w, r)
			return
		}
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

		dashboardUser, err := h.daService.GetValidatorDashboardUser(r.Context(), types.VDBIdPrimary(dashboardId))
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

// Common middleware logic for checking user premium perks
func (h *HandlerService) PremiumPerkCheckMiddleware(next http.Handler, hasRequiredPerk func(premiumPerks types.PremiumPerks) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get user id from context
		userId, err := GetUserIdByContext(r)
		if err != nil {
			handleErr(w, r, err)
			return
		}

		// get user info
		userInfo, err := h.daService.GetUserInfo(r.Context(), userId)
		if err != nil {
			handleErr(w, r, err)
			return
		}

		// check if user has the required premium perk
		if !hasRequiredPerk(userInfo.PremiumPerks) {
			handleErr(w, r, newForbiddenErr("users premium perks do not allow usage of this endpoint"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Middleware for managing dashboards via API
func (h *HandlerService) ManageDashboardsViaApiCheckMiddleware(next http.Handler) http.Handler {
	return h.PremiumPerkCheckMiddleware(next, func(premiumPerks types.PremiumPerks) bool {
		return premiumPerks.ManageDashboardViaApi
	})
}

// Middleware for managing notifications via API
func (h *HandlerService) ManageNotificationsViaApiCheckMiddleware(next http.Handler) http.Handler {
	return h.PremiumPerkCheckMiddleware(next, func(premiumPerks types.PremiumPerks) bool {
		return premiumPerks.ConfigureNotificationsViaApi
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
		dashboard, err := h.daService.GetValidatorDashboardInfo(r.Context(), dashboardId.Id)
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

// middleware that checks if the request wants mocked data and if the user is allowed to use it. the flag is stored in the request context.
// note that mocked data is only returned by handlers that support it.
func (h *HandlerService) SetIsMockedFlagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isMocked, _ := strconv.ParseBool(r.Header.Get("is_mocked"))
		if !isMocked {
			next.ServeHTTP(w, r)
			return
		}
		// fetch user group
		userId, err := h.GetUserIdBySession(r)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		userCredentials, err := h.daService.GetUserInfo(r.Context(), userId)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		allowedGroups := []string{types.UserGroupAdmin, types.UserGroupDev}
		if !slices.Contains(allowedGroups, userCredentials.UserGroup) {
			handleErr(w, r, newForbiddenErr("user is not allowed to use mock data"))
			return
		}
		// store isMocked flag in context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxIsMockedKey, true)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
