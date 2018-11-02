package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/json-iterator/go" // I chose this library becase it is faster than the standard one
)

const (
	maxLen  = 100  //Maximum ID length
	maxSize = 1024 //Maximum content length (bytes)
)

var (
	db *bolt.DB
)

func main() {
	var err error

	db, err = bolt.Open("objects.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("objects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return err
	})
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Route("/api/objects", func(r chi.Router) {
		r.Get("/", listObjects)
		r.NotFound(wrongID)
		r.Route("/{objectID:[a-zA-Z0-9]+}", func(r chi.Router) {
			r.Get("/", getObject)
			r.Put("/", updateObject)
			r.Delete("/", delObject)
		})
	})

	http.ListenAndServe(":8080", r)
}

type Object struct {
	ContentType string `json:"Content-Type"`
	Data        string `json:"Data"`
}

func listObjects(w http.ResponseWriter, r *http.Request) {
	keys := dbGetKeys([]byte("objects"))
	if len(keys) == 0 {
		w.Write([]byte("[]"))
		return
	}
	resp, err := jsoniter.Marshal(keys)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Data cannot be processed"))
		return
	}
	w.Write(resp)
}

func updateObject(w http.ResponseWriter, r *http.Request) {
	var buff []byte
	var obj Object
	var err error

	// Syntax check,and limits checking
	objectID := chi.URLParam(r, "objectID")
	if len(objectID) > 100 {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Incorrect ID: identifier exceeds %v characters", maxLen)))
		return
	} else if r.ContentLength > 1024 {
		w.WriteHeader(413)
		w.Write([]byte(fmt.Sprintf("Content cannot exceeds %v bytes", maxSize)))
		return
	} else if obj.ContentType = r.Header.Get("Content-Type"); obj.ContentType == "" {
		w.WriteHeader(400)
		w.Write([]byte("Content-Type cannot be empty"))
		return
	}

	// Read body request and convert to an json format object
	buff, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Request body read failed with '%s'\n", err)
		w.WriteHeader(500)
		w.Write([]byte("Request processing error"))
		return
	}
	obj.Data = string(buff)
	json, err := jsoniter.Marshal(obj)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Data cannot be processed"))
		return
	}

	// Store in database
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("objects"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte(objectID), json)
		return err
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Database temporary error"))
		return
	}
	w.WriteHeader(201)
}

func getObject(w http.ResponseWriter, r *http.Request) {
	if resp := dbGet([]byte("objects"), []byte(chi.URLParam(r, "objectID"))); len(resp) > 0 {
		w.Write(resp)
		return
	}
	w.WriteHeader(404)
	w.Write([]byte(fmt.Sprintf("%v does not exist", chi.URLParam(r, "objectID"))))
}

func delObject(w http.ResponseWriter, r *http.Request) {
	if err := dbDel([]byte("objects"), []byte(chi.URLParam(r, "objectID"))); err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
	}
}

//												Database operations

func dbGet(bucketName, key []byte) (data []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data = b.Get(key)
		return nil
	})
	return data
}

func dbDel(bucketName, key []byte) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if len(b.Get(key)) != 0 {
			return b.Delete(key)
		}
		return errors.New("Record does not exist")
	})
}

func dbGetKeys(bucketName []byte) (list []string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		b.ForEach(func(k, v []byte) error {
			list = append(list, string(k))
			return nil
		})
		return nil
	})
	return list
}

//												Error handling functions

func wrongID(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	w.Write([]byte("Incorrect ID: wrong ID syntax"))
}
