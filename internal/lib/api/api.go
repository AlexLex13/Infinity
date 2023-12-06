package api

import (
	"net/http"
)

func GetRedirect(url string) (string, error) {

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // stop after 1st redirect
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		return "", nil
	}

	return resp.Header.Get("Location"), nil
}
