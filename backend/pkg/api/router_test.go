package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlagToggle(t *testing.T) {
	tests := []struct {
		name             string
		requestFeature   bool
		isFeatureAllowed bool
		expectFeature    bool
	}{
		{
			name:             "Feature requested and allowed",
			requestFeature:   true,
			isFeatureAllowed: true,
			expectFeature:    true,
		},
		{
			name:             "Feature requested but not allowed",
			requestFeature:   true,
			isFeatureAllowed: false,
			expectFeature:    false,
		},
		{
			name:             "Feature not requested and allowed",
			requestFeature:   false,
			isFeatureAllowed: true,
			expectFeature:    false,
		},
		{
			name:             "Feature not requested and not allowed",
			requestFeature:   false,
			isFeatureAllowed: false,
			expectFeature:    false,
		},
	}

	const featureFlag = "new_feature"
	featureHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	legacyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}
	for _, tt := range tests {
		isFeatureAllowedFunc := func(string) bool {
			return tt.isFeatureAllowed
		}
		t.Run(tt.name, func(t *testing.T) {
			var queryParam string
			if tt.requestFeature {
				queryParam = "?feature_flags=" + featureFlag
			}
			req := httptest.NewRequest(http.MethodGet, "/"+queryParam, nil)
			w := httptest.NewRecorder()

			handler := featureFlagToggle(isFeatureAllowedFunc, featureFlag, legacyHandler, featureHandler)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if tt.expectFeature {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			}
		})
	}
}
