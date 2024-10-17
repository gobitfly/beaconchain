package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
)

// Middlewares

func hashUint64(data uint64) [32]byte {
	// Convert uint64 to a byte slice
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, data)

	// Compute SHA-256 hash
	hash := sha256.Sum256(buf)
	return hash
}

func checkHash(data uint64, hashStr string) bool {
	// Decode the hexadecimal string into a byte slice
	hashToCheck, err := hex.DecodeString(hashStr)
	if err != nil {
		return false
	}

	// Hash the uint64 value
	computedHash := hashUint64(data)

	// Compare the computed hash with the provided hash
	return string(computedHash[:]) == string(hashToCheck)
}

// returns a middleware that stores user id in context, using the provided function
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

		// if user id matches a given hash, allow access without checking dashboard access and return mock data
		// TODO: move to config, exposing this in source code is a minor security risk for now
		validHashes := []string{
			"2cab06069254b5555b617efa1d17f0748324270bb587b73422e6840d59ff322c",
			"fc624cf355b84bc583661552982894621568b59c0a1c92ab0c1e03ed3bbf649b",
			"03e7fb02cbc33eb45e98ab50b4bcad7fc338e5edfb5eca33ad9eb7d13d4ff106",
		}
		for _, hash := range validHashes {
			if checkHash(userId, hash) {
				ctx := r.Context()
				ctx = context.WithValue(ctx, ctxIsMockEnabledKey, true)
				r = r.WithContext(ctx)
			}
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserIdKey, userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (h *HandlerService) StoreUserIdBySessionMiddleware(next http.Handler) http.Handler {
	return StoreUserIdMiddleware(next, func(r *http.Request) (uint64, error) {
		return h.GetUserIdBySession(r)
	})
}

func (h *HandlerService) StoreUserIdByApiKeyMiddleware(next http.Handler) http.Handler {
	return StoreUserIdMiddleware(next, func(r *http.Request) (uint64, error) {
		return h.GetUserIdByApiKey(r)
	})
}

// returns a middleware that checks if user has access to dashboard when a primary id is used
func (h *HandlerService) VDBAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if mock data is used, no need to check access
		if isMockEnabled, ok := r.Context().Value(ctxIsMockEnabledKey).(bool); ok && isMockEnabled {
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
		userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
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
