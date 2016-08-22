package spaassets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestDeepLink(t *testing.T) {

	c.Convey("Given an endpoint with DeepLink middleware", t, func() {
		success := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		c.Convey("configured with appURL=/app", func() {

			handler := DeepLink(success, "/app/")

			c.Convey("request to /app/config.js should not rewrite URL", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				c.So(r.URL.Path, c.ShouldEqual, "/app/config.js")
			})

			c.Convey("request to /app/home should be rewritten to /app/", func() {

				r, _ := http.NewRequest("GET", "http://host/app/home", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				c.So(r.URL.Path, c.ShouldEqual, "/app/")
			})

			c.Convey("request to /app/home/ should be rewritten to /app/", func() {

				r, _ := http.NewRequest("GET", "http://host/app/home/", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				c.So(r.URL.Path, c.ShouldEqual, "/app/")
			})
		})
	})
}

func TestMaxAgeCacheHandler(t *testing.T) {

	c.Convey("Given an endpoint with MaxAgeCacheHandler middleware", t, func() {

		success := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		c.Convey("configured with age=0, include=true, prefix=/app", func() {

			handler := MaxAgeCacheHandler(success, 0, true, "/app")

			c.Convey("request to /app/config.js should set cache header to max-age=0", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "max-age=0")
			})

			c.Convey("request to /config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})
		})

		c.Convey("configured with age=10, include=false, prefix=/app", func() {

			handler := MaxAgeCacheHandler(success, 10, false, "/app")

			c.Convey("request to /app/config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})

			c.Convey("request to /config.js should set cache header to max-age=0", func() {

				r, _ := http.NewRequest("GET", "http://host/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "max-age=10")
			})
		})

		c.Convey("configured with age=0, include=true, prefix=", func() {

			handler := MaxAgeCacheHandler(success, 0, true, "")

			c.Convey("request to /app/config.js should set cache header to max-age=0", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "max-age=0")
			})
		})

		c.Convey("configured with age=10, include=false, prefix=", func() {

			handler := MaxAgeCacheHandler(success, 10, false, "")

			c.Convey("request to /app/config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})
		})
	})
}

func TestNoStoreCacheHandler(t *testing.T) {

	c.Convey("Given an endpoint with NoStoreCacheHandler middleware", t, func() {

		success := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})

		c.Convey("configured with include=true, prefix=/app", func() {

			handler := NoStoreCacheHandler(success, true, "/app")

			c.Convey("request to /app/config.js should set cache header to no-store", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "no-store")
			})

			c.Convey("request to /config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})
		})

		c.Convey("configured with include=false, prefix=/app", func() {

			handler := NoStoreCacheHandler(success, false, "/app")

			c.Convey("request to /app/config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})

			c.Convey("request to /config.js should set cache header to no-store", func() {

				r, _ := http.NewRequest("GET", "http://host/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "no-store")
			})
		})

		c.Convey("configured with include=true, prefix=", func() {

			handler := NoStoreCacheHandler(success, true, "")

			c.Convey("request to /app/config.js should set cache header to no-store", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldContainSubstring, "no-store")
			})
		})

		c.Convey("configured with include=false, prefix=", func() {

			handler := NoStoreCacheHandler(success, false, "")

			c.Convey("request to /app/config.js should not set cache header", func() {

				r, _ := http.NewRequest("GET", "http://host/app/config.js", nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)

				header := w.Header().Get("Cache-Control")

				c.So(header, c.ShouldBeEmpty)
			})
		})
	})
}
