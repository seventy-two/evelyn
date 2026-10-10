package quotes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seventy-two/evelyn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetQuoteOfTheDay(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"q":"Ability is a poor man's wealth.","a":"John Wooden","i":"","h":"","date":"2026-10-10"}]`))
	}))
	defer srv.Close()
	serviceConfig = &service.Service{TargetURL: srv.URL}

	q, err := getQuoteOfTheDay()
	require.NoError(t, err)
	assert.Equal(t, "Ability is a poor man's wealth.", q.quote)
	assert.Equal(t, "John Wooden", q.author)
}

func TestGetQuoteOfTheDayEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	serviceConfig = &service.Service{TargetURL: srv.URL}

	_, err := getQuoteOfTheDay()
	assert.Error(t, err)
}
