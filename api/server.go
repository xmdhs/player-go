package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/xmdhs/player-go/utils"
	"go.etcd.io/bbolt"
)

func Server(cxt context.Context, db *bbolt.DB) int {
	p := utils.GetProt()

	mux := httprouter.New()

	mux.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", "*")
			header.Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
			header.Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.POST("/store/:key", store(db))
	mux.GET("/read/:key", read(db))

	s := http.Server{
		Addr:              "127.0.0.1:" + strconv.FormatInt(p, 10),
		Handler:           mux,
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
