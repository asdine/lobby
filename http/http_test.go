package http_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asdine/lobby"
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

	t.Run("InvalidJSON", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`hello`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BackendNotFound", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return lobby.ErrBackendNotFound
		}

		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   bucket   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("BackendConflict", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return lobby.ErrBucketAlreadyExists
		}

		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   bucket   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return errors.New("something unexpected happened !")
		}

		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   bucket   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, bucketName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "bucket", bucketName)

			return nil
		}

		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   bucket   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestSaveItem(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/v1/b/bucket/key", bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, lobby.ErrBucketNotFound
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/v1/b/bucket/key", strings.NewReader(`{}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, errors.New("something unexpected happened !")
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/v1/b/bucket/key", strings.NewReader(`{}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry)

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				SaveFn: func(key string, data []byte) (*lobby.Item, error) {
					require.Equal(t, "key", key)
					require.Equal(t, []byte(`"hello"`), data)

					return &lobby.Item{
						Key:  key,
						Data: data,
					}, nil
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/v1/b/bucket/key", strings.NewReader(`hello`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
