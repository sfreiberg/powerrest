package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"

	"database/sql"
	"log"
	"net/http"
)

var (
	db   *sql.DB
	conf *Config
)

func main() {
	conf, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open(conf.DbType, conf.DbConn)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/v1/domains", domainIndex).Methods("GET")                 // list domains
	router.HandleFunc("/v1/domains", domainCreate).Methods("POST")               // new domain
	router.HandleFunc("/v1/domains/{id:[0-9]+}", domainShow).Methods("GET")      // show domain
	router.HandleFunc("/v1/domains/{id:[0-9]+}", domainUpdate).Methods("POST")   // update domain
	router.HandleFunc("/v1/domains/{id:[0-9]+}", domainDelete).Methods("DELETE") // delete domain

	router.HandleFunc("/v1/records", recordIndex).Methods("GET")                 // list domains
	router.HandleFunc("/v1/records", recordCreate).Methods("POST")               // new domain
	router.HandleFunc("/v1/records/{id:[0-9]+}", recordShow).Methods("GET")      // show domain
	router.HandleFunc("/v1/records/{id:[0-9]+}", recordUpdate).Methods("POST")   // update domain
	router.HandleFunc("/v1/records/{id:[0-9]+}", recordDelete).Methods("DELETE") // delete domain

	log.Println("Listening on", conf.ListenAddr)
	log.Fatal(http.ListenAndServe(conf.ListenAddr, router))
}
