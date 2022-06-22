package cors

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xmdhs/player-go/utils"
)

func Server(cxt context.Context, t *http.Transport) int {
	p := utils.GetProt()
	s := http.Server{
		Addr:              "127.0.0.1:" + strconv.FormatInt(p, 10),
		Handler:           handler(t),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go s.ListenAndServe()
	go func() {
		<-cxt.Done()
		s.Shutdown(cxt)
	}()
	return int(p)
}

func handler(t *http.Transport) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		proxyURL := q.Get("_proxyURL")
		if proxyURL != "" {
			purl, err := url.Parse(proxyURL)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			corsProxy(purl, t, q.Get("_referer")).ServeHTTP(w, r)
			return
		}
		u := r.URL.String()
		u = strings.TrimPrefix(u, "/")
		purl, err := url.Parse(u)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if purl.Scheme == "" {
			purl.Scheme = "http"
		}
		if purl.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		corsProxy(purl, t, "").ServeHTTP(w, r)
	}
}

func corsProxy(u *url.URL, t *http.Transport, referer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Host:   u.Host,
			Scheme: u.Scheme,
		})
		proxy.ErrorLog = log.Default()
		proxy.Transport = t

		df := proxy.Director

		proxy.Director = func(r *http.Request) {
			df(r)
			r.Header.Del("referer")
			r.Header.Del("origin")
			r.Header.Del("X-Forwarded-For")
			r.Header.Del("X-Real-IP")
			r.Host = u.Host
			if referer != "" {
				r.Header.Set("referer", referer)
			}
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			r.Header.Set("Access-Control-Allow-Origin", "*")
			r.Header.Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
			r.Header.Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
			r.Header.Set("X-ToProxy", r.Request.URL.String())
			if r.StatusCode >= 300 && r.StatusCode < 400 && r.Header.Get("Location") != "" {
				r.Header.Set("Location", "/"+r.Header.Get("Location"))
			}
			return nil
		}

		r.URL = u
		r.RemoteAddr = ""
		r.RequestURI = u.String()

		proxy.ServeHTTP(w, r)
	}
}
