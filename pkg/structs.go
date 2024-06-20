package pkg

import "net/http"

type URLShortened struct {
	Key string `json:"key" gorm:"primaryKey;unique"`

	Origin  string  `json:"origin"`
	WebHook *string `json:"webhook,omitempty"`
}

type WebHookRequest struct {
	http.Header `json:"header"`
}

type WebHookResponse struct {
	Pass   bool   `json:"pass"`
	Reason string `json:"reason"`
}
