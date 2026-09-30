package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		origin     string
		host       string
		wantStatus int
	}{
		{name: "post same origin passes", method: http.MethodPost, origin: "http://localhost:8080", host: "localhost:8080", wantStatus: http.StatusOK},
		{name: "post without origin passes", method: http.MethodPost, origin: "", host: "localhost:8080", wantStatus: http.StatusOK},
		{name: "post cross origin rejected", method: http.MethodPost, origin: "http://evil.example", host: "localhost:8080", wantStatus: http.StatusForbidden},
		{name: "get cross origin passes", method: http.MethodGet, origin: "http://evil.example", host: "localhost:8080", wantStatus: http.StatusOK},
		{name: "head cross origin passes", method: http.MethodHead, origin: "http://evil.example", host: "localhost:8080", wantStatus: http.StatusOK},
		{name: "put cross origin rejected", method: http.MethodPut, origin: "https://evil.example", host: "localhost:8080", wantStatus: http.StatusForbidden},
		{name: "delete cross origin rejected", method: http.MethodDelete, origin: "http://evil.example:1234", host: "localhost:8080", wantStatus: http.StatusForbidden},
		{name: "malformed origin rejected", method: http.MethodPost, origin: "http://[::1", host: "localhost:8080", wantStatus: http.StatusForbidden},
		{name: "empty host origin rejected", method: http.MethodPost, origin: "http://", host: "localhost:8080", wantStatus: http.StatusForbidden},
		{name: "host comparison is case insensitive", method: http.MethodPost, origin: "http://LOCALHOST:8080", host: "localhost:8080", wantStatus: http.StatusOK},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := CheckOrigin(next)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/add_model", nil)
			req.Host = tc.host
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("method %s origin %q host %q: expected status %d, got %d", tc.method, tc.origin, tc.host, tc.wantStatus, rr.Code)
			}
		})
	}
}
