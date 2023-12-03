package save

import resp "github.com/AlexLex13/Infinity/internal/lib/api/response"

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}
