package redirect_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexLex13/Infinity/internal/http-server/handlers/redirect"
	"github.com/AlexLex13/Infinity/internal/http-server/handlers/redirect/mocks"
	"github.com/AlexLex13/Infinity/internal/lib/api/response"
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
			name:      "Non existent alias",
			alias:     "non_existent_alias",
			respError: "not found",
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "Redirect error",
			alias:     "test_alias",
			respError: "internal error",
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

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse // stop after 1st redirect
				},
			}

			resp, err := client.Get(ts.URL + "/" + tc.alias)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusFound {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var res response.Response

				require.NoError(t, json.Unmarshal(body, &res))

				require.Equal(t, tc.respError, res.Error)
			} else {
				assert.Equal(t, tc.url, resp.Header.Get("Location"))
			}
		})
	}
}
