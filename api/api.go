package api

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.etcd.io/bbolt"
)

func store(db *bbolt.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		key := p.ByName("key")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = db.Update(func(tx *bbolt.Tx) error {
			bkt, err := tx.CreateBucketIfNotExists([]byte("data"))
			if err != nil {
				return err
			}
			return bkt.Put([]byte(key), b)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func read(db *bbolt.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		key := p.ByName("key")
		err := db.View(func(tx *bbolt.Tx) error {
			bkt := tx.Bucket([]byte("data"))
			if bkt == nil {
				http.NotFound(w, r)
				return nil
			}
			v := bkt.Get([]byte(key))
			if v == nil {
				http.NotFound(w, r)
				return nil
			}
			w.Write(v)
			return nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
