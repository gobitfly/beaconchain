package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware(t *testing.T) {
	// Create a dummy handler that does not panic
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	})

	// Wrap with RecoveryMiddleware
	middleware := RecoveryMiddleware()
	wrappedNormal := middleware(normalHandler)

	t.Run("normal handler does not panic", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		wrappedNormal.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "ok", rec.Body.String())
	})

	t.Run("panic handler is recovered", func(t *testing.T) {
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something bad happened")
		})

		wrappedPanic := middleware(panicHandler)
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		// Should not panic
		require.NotPanics(t, func() {
			wrappedPanic.ServeHTTP(rec, req)
		})

		require.Equal(t, http.StatusInternalServerError, rec.Code)
		require.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rec.Body.String())
	})
}
