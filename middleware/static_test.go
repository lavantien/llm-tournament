package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestStaticFiles(t *testing.T) {
	fs := fstest.MapFS{
		"output.css":         {Data: []byte("body{}")},
		"utils.js":           {Data: []byte("function f(){}")},
		"logo.webp":          {Data: []byte("webp")},
		"shared.go":          {Data: []byte("package templates")},
		"prompt_list.html":   {Data: []byte("<html></html>")},
		"shared_test.go":     {Data: []byte("package templates")},
		"sub/dir/styles.css": {Data: []byte("a{}")},
	}
	handler := http.StripPrefix("/static/", StaticFiles(http.FS(fs)))

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "css served", method: http.MethodGet, path: "/static/output.css", wantStatus: http.StatusOK},
		{name: "js served", method: http.MethodGet, path: "/static/utils.js", wantStatus: http.StatusOK},
		{name: "webp served", method: http.MethodGet, path: "/static/logo.webp", wantStatus: http.StatusOK},
		{name: "nested css served", method: http.MethodGet, path: "/static/sub/dir/styles.css", wantStatus: http.StatusOK},
		{name: "go source blocked", method: http.MethodGet, path: "/static/shared.go", wantStatus: http.StatusNotFound},
		{name: "html template blocked", method: http.MethodGet, path: "/static/prompt_list.html", wantStatus: http.StatusNotFound},
		{name: "test source blocked", method: http.MethodGet, path: "/static/shared_test.go", wantStatus: http.StatusNotFound},
		{name: "directory blocked", method: http.MethodGet, path: "/static/sub/", wantStatus: http.StatusNotFound},
		{name: "missing file stays 404", method: http.MethodGet, path: "/static/nope.css", wantStatus: http.StatusNotFound},
		{name: "post blocked", method: http.MethodPost, path: "/static/output.css", wantStatus: http.StatusNotFound},
		{name: "head allowed", method: http.MethodHead, path: "/static/output.css", wantStatus: http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("%s %s: expected status %d, got %d", tc.method, tc.path, tc.wantStatus, rr.Code)
			}
		})
	}
}
