package remove_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexLex13/Infinity/internal/http-server/handlers/remove"
	"github.com/AlexLex13/Infinity/internal/http-server/handlers/remove/mocks"
	"github.com/AlexLex13/Infinity/internal/lib/api/response"
	"github.com/AlexLex13/Infinity/internal/lib/logger/handlers/slogdiscard"
	"github.com/AlexLex13/Infinity/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestRemoveHandler(t *testing.T) {
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
			name:      "Delete error",
			alias:     "test_alias",
			respError: "internal error",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Delete("/{alias}", remove.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			client := &http.Client{}

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tc.alias, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNoContent {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var res response.Response

				require.NoError(t, json.Unmarshal(body, &res))

				require.Equal(t, tc.respError, res.Error)
			}
		})
	}
}
