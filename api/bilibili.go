package api

import (
	"encoding/json"
	"io"
	"net/http"
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
		reqs, err := http.NewRequest("GET", "https://api.bilibili.com/x/player/playurl?"+uq.Encode(), nil)
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
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bvideo := bili[biliVideoInfo]{}
		err = json.Unmarshal(b, &bvideo)
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
		w.Write([]byte(bvideo.Data.Durl[0].URL))
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
