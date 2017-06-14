package http_test

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`hello`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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

		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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

		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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

		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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

		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/buckets/backend", strings.NewReader(`{"name": "   bucket   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestSaveItem(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("PUT", "/v1/b/bucket/key", bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

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
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				PutFn: func(key string, value []byte) (*lobby.Item, error) {
					require.Equal(t, "key", key)
					require.Equal(t, []byte(`"hello"`), value)

					return &lobby.Item{
						Key:   key,
						Value: value,
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

func TestGetItem(t *testing.T) {
	t.Run("BucketNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, lobby.ErrBucketNotFound
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					require.Equal(t, "key", key)

					return nil, lobby.ErrKeyNotFound
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				GetFn: func(key string) (*lobby.Item, error) {
					require.Equal(t, "key", key)

					return &lobby.Item{
						Key:   key,
						Value: []byte(`"hello"`),
					}, nil
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, `"hello"`, w.Body.String())
	})
}

func TestDeleteItem(t *testing.T) {
	t.Run("BucketNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return nil, lobby.ErrBucketNotFound
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("DELETE", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					require.Equal(t, "key", key)

					return lobby.ErrKeyNotFound
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("DELETE", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(os.Stderr, "", log.LstdFlags))

		registry.BucketFn = func(name string) (lobby.Bucket, error) {
			require.Equal(t, "bucket", name)

			return &mock.Bucket{
				DeleteFn: func(key string) error {
					require.Equal(t, "key", key)

					return nil
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("DELETE", "/v1/b/bucket/key", nil)
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNoContent, w.Code)
	})
}
