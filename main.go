package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"github.com/syndtr/goleveldb/leveldb"

	_ "embed"
)

var db *leveldb.DB

func main() {
	var err error
	db, err = leveldb.OpenFile("data", nil)
	if err != nil {
		log.Println(err)
		db, err = leveldb.RecoverFile("data", nil)
		if err != nil {
			panic(err)
		}
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/keys", listKeys)
	mux.HandleFunc("/getvalue", getValue)
	mux.HandleFunc("/", index)

	s := http.Server{
		Addr:              "127.2.32.2:18080",
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	browser.OpenURL("http://" + s.Addr)
	log.Println(s.ListenAndServe())
}

//go:embed index.html
var indexhtml []byte

func index(w http.ResponseWriter, req *http.Request) {
	w.Write(indexhtml)
}

func listKeys(w http.ResponseWriter, req *http.Request) {
	keys := []rdata{}
	itr := db.NewIterator(nil, nil)
	defer itr.Release()
	for itr.Next() {
		keys = append(keys, rdata{
			Value:  string(itr.Key()),
			Base64: base64.StdEncoding.EncodeToString(itr.Key()),
		})
	}
	b, err := json.Marshal(keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func getValue(w http.ResponseWriter, req *http.Request) {
	key := req.FormValue("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}
	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := db.Get(k, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	s := base64.StdEncoding.EncodeToString(b)
	r := rdata{
		Value:  string(b),
		Base64: s,
	}
	bb, err := json.Marshal(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bb)
}

type rdata struct {
	Value  string `json:"value"`
	Base64 string `json:"base64"`
}
