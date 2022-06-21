package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
)

func bilivideoGet(t *http.Transport) httprouter.Handle {
	c := http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
	}
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		qn := r.FormValue("qn")
		if qn == "" {
			qn = "120"
		}
		bvid := r.FormValue("bvid")
		cid := r.FormValue("cid")

		if bvid == "" || cid == "" {
			http.Error(w, "bvid or cid is required", http.StatusBadRequest)
			return
		}
		fnval := "0"
		if qn == "120" {
			fnval = "128"
		}
		DedeUserID := r.FormValue("DedeUserID")
		DedeUserID__ckMd5 := r.FormValue("DedeUserID__ckMd5")
		SESSDATA := r.FormValue("SESSDATA")
		bili_jct := r.FormValue("bili_jct")

		uq := url.Values{}
		uq.Set("bvid", bvid)
		uq.Set("cid", cid)
		uq.Set("fnval", fnval)
		uq.Set("qn", qn)
		reqs, err := http.NewRequest("GET", "https://api.bilibili.com/x/player/playurl"+uq.Encode(), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reqs.Header.Set("cookie", "DedeUserID="+DedeUserID+"; DedeUserID__ckMd5="+DedeUserID__ckMd5+"; SESSDATA="+SESSDATA+"; bili_jct="+bili_jct)
		resp, err := c.Do(reqs)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bvideo := bili[biliVideoInfo]{}
		err = json.NewDecoder(resp.Body).Decode(&bvideo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if bvideo.Code != 0 {
			http.Error(w, bvideo.Message, http.StatusInternalServerError)
			return
		}
		if len(bvideo.Data.Durl) != 1 {
			http.Error(w, "无法处理分段视频", http.StatusInternalServerError)
			return
		}
		// flv
		w.Header().Set("Content-Type", "video/x-flv")
		proxyHandler(bvideo.Data.Durl[0].URL, t).ServeHTTP(w, r)
	}
}

type bili[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
}

type biliVideoInfo struct {
	AcceptDescription []string        `json:"accept_description"`
	AcceptFormat      string          `json:"accept_format"`
	AcceptQuality     []int           `json:"accept_quality"`
	Durl              []biliVideoDurl `json:"durl"`
	Format            string          `json:"format"`
	From              string          `json:"from"`
	Message           string          `json:"message"`
	Quality           int             `json:"quality"`
	Result            string          `json:"result"`
	SeekParam         string          `json:"seek_param"`
	SeekType          string          `json:"seek_type"`
	Timelength        int             `json:"timelength"`
	VideoCodecid      int             `json:"video_codecid"`
}

type biliVideoDurl struct {
	Ahead     string   `json:"ahead"`
	BackupURL []string `json:"backup_url"`
	Length    int      `json:"length"`
	Order     int      `json:"order"`
	Size      int      `json:"size"`
	URL       string   `json:"url"`
	Vhead     string   `json:"vhead"`
}

func proxyHandler(u string, t *http.Transport) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		corsProxy(purl, t).ServeHTTP(w, r)
	}
}

func corsProxy(u *url.URL, t *http.Transport) http.HandlerFunc {
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
			r.Header.Set("referer", "https://www.bilibili.com")
			r.Host = u.Host
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			r.Header.Set("Access-Control-Allow-Origin", "*")
			r.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			r.Header.Set("Access-Control-Allow-Headers", "Accept, Authorization, Cache-Control, Content-Type, DNT, If-Modified-Since, Keep-Alive, Origin, User-Agent")
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
