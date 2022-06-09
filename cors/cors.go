package cors

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Server(cxt context.Context) int {
	p := getProt()
	s := http.Server{
		Addr:              "127.0.0.1:" + strconv.FormatInt(p, 10),
		Handler:           handler(),
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

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		corsProxy(purl).ServeHTTP(w, r)
	}
}

func corsProxy(u *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Host:   u.Host,
			Scheme: u.Scheme,
		})
		proxy.ErrorLog = log.Default()

		df := proxy.Director

		proxy.Director = func(r *http.Request) {
			df(r)
			r.Header.Del("referer")
			r.Header.Del("origin")
			r.Header.Del("X-Forwarded-For")
			r.Header.Del("X-Real-IP")
			r.Host = u.Host
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			r.Header.Set("Access-Control-Allow-Origin", r.Header.Get("origin"))
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

func getProt() int64 {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	list := strings.Split(l.Addr().String(), ":")
	i, err := strconv.ParseInt(list[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}
