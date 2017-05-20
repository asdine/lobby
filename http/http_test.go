package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	lobbyHttp "github.com/asdine/lobby/http"
	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBucket(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}
