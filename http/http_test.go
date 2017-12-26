package http_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asdine/lobby"
	lobbyHttp "github.com/asdine/lobby/http"
	"github.com/asdine/lobby/log"
	"github.com/asdine/lobby/mock"
	"github.com/stretchr/testify/require"
)

func createTopicRequest(t *testing.T, r io.Reader) *http.Request {
	req, err := http.NewRequest("POST", "/v1/topics", r)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestCreateTopic(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`hello`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidationError", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`{"name": "   "}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BackendNotFound", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, topicName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "topic", topicName)

			return lobby.ErrBackendNotFound
		}

		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`{"name": "   topic   ", "backend": "backend"}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("BackendConflict", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, topicName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "topic", topicName)

			return lobby.ErrTopicAlreadyExists
		}

		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`{"name": "   topic   ","backend": "backend"}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, topicName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "topic", topicName)

			return errors.New("something unexpected happened !")
		}

		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`{"name": "   topic   ", "backend": "backend"}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry

		registry.CreateFn = func(backendName, topicName string) error {
			require.Equal(t, "backend", backendName)
			require.Equal(t, "topic", topicName)

			return nil
		}

		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r := createTopicRequest(t, strings.NewReader(`{"name": "   topic   ", "backend": "backend"}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestSaveMessage(t *testing.T) {
	t.Run("EmptyBody", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/topics/topic/key", bytes.NewReader([]byte(nil)))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("TopicNotFound", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		registry.TopicFn = func(name string) (lobby.Topic, error) {
			require.Equal(t, "topic", name)

			return nil, lobby.ErrTopicNotFound
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/topics/topic/key", strings.NewReader(`{}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InternalError", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		registry.TopicFn = func(name string) (lobby.Topic, error) {
			require.Equal(t, "topic", name)

			return nil, errors.New("something unexpected happened !")
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/topics/topic/key", strings.NewReader(`{}`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("OK", func(t *testing.T) {
		var registry mock.Registry
		h := lobbyHttp.NewHandler(&registry, log.New(log.Output(ioutil.Discard)))

		registry.TopicFn = func(name string) (lobby.Topic, error) {
			require.Equal(t, "topic", name)

			return &mock.Topic{
				SendFn: func(message *lobby.Message) error {
					require.Equal(t, "group", message.Group)
					require.Equal(t, []byte(`hello`), message.Value)

					return nil
				},
			}, nil
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/v1/topics/topic/group", strings.NewReader(`hello`))
		h.ServeHTTP(w, r)
		require.Equal(t, http.StatusCreated, w.Code)
	})
}
