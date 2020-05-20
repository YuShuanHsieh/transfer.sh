package apiauth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v5"
	apiauth "github.com/dutchcoders/transfer.sh/api-auth"
	"github.com/stretchr/testify/assert"
)

func TestAPIAuthenticator(t *testing.T) {
	headerKey := "AUTH-Header"
	headerValue := gofakeit.Word()
	user := gofakeit.Word()
	password := gofakeit.Word()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get(headerKey)
		if headerValue == value {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	t.Run("valid", func(t *testing.T) {
		cfg := apiauth.APIConfig{
			Endpoint: ts.URL,
			Headers:  map[string]string{headerKey: headerValue},
		}
		api := apiauth.New(cfg)
		result, err := api.Authenticate(user, password)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("invalid", func(t *testing.T) {
		cfg := apiauth.APIConfig{
			Endpoint: ts.URL,
			Headers:  nil,
		}
		api := apiauth.New(cfg)
		result, err := api.Authenticate(user, password)
		assert.Error(t, err)
		assert.False(t, result)
	})
}
