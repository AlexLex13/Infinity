package redirect_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/AlexLex13/Infinity/internal/http-server/handlers/redirect"
	"github.com/AlexLex13/Infinity/internal/http-server/handlers/redirect/mocks"
	"github.com/AlexLex13/Infinity/internal/lib/api"
	"github.com/AlexLex13/Infinity/internal/lib/logger/handlers/slogdiscard"
	"github.com/AlexLex13/Infinity/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "With exist alias",
			alias: "test_alias",
			url:   "https://www.google.com/",
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "invalid request",
		},
		{
			name:      "Non existent alias",
			alias:     "non_existent_alias",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "Redirect error",
			alias:     "test_alias",
			respError: "failed to redirect",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			assert.Equal(t, tc.url, redirectedToURL)
		})
	}
}
