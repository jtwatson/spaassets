package webapps

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// DeepLink middleware handles deep-linking into a single page application (SPA).
//
// This middleware should wrap your asset handler, and expects all requests are
// requesting a file with an extention. (ie: .js, .html, .png, .jpg etc)
//
// If it recieves a request targeting a directory (with or without a trailing '/'),
// it will rewrite the request to target appURL
//
// appURL should be the location of the SPA
//
func DeepLink(next http.Handler, appURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ext := filepath.Ext(r.URL.Path); ext == "" {
			r.URL.Path = appURL
		}
		next.ServeHTTP(w, r)
	})
}

// NoStoreCacheHandler applies Cache-control: no-store header according to options.
//
// Options
// include: indicates if matches should be included/excluded.
// prefix:  prefix is matched to the beginning of url.
//
func NoStoreCacheHandler(next http.Handler, include bool, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, prefix) == include {
			w.Header().Set("Cache-control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// MaxAgeCacheHandler applies Cache-control: max-age=<age> header according to options.
//
// Options
// age: cache experation in seconds
// include: indicates if matches should be included/excluded.
// prefix:  prefix is matched to the beginning of url.
//
func MaxAgeCacheHandler(next http.Handler, age int, include bool, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, prefix) == include {
			w.Header().Set("Cache-control", fmt.Sprintf("max-age=%d", age))
		}
		next.ServeHTTP(w, r)
	})
}
