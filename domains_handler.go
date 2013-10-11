package main

import (
	"github.com/gorilla/mux"

	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func domainIndex(w http.ResponseWriter, r *http.Request) {
	domains := AllDomains()

	enc := json.NewEncoder(w)
	err := enc.Encode(domains)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func domainCreate(w http.ResponseWriter, r *http.Request) {
	domain := NewDomain("")

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = domain.Create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func domainShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	d, err := FindDomain(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r) // return a 404
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func domainUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	domain, err := FindDomain(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	err = jsonDecoder.Decode(&domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = domain.Update()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func domainDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	domain := &Domain{Id: id}
	err = domain.Delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
