package main

import (
	"github.com/gorilla/mux"

	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

func recordIndex(w http.ResponseWriter, r *http.Request) {
	records := AllRecords()

	enc := json.NewEncoder(w)
	err := enc.Encode(records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recordCreate(w http.ResponseWriter, r *http.Request) {
	record := &Record{}

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = record.Create()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recordShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	record, err := FindRecord(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r) // return a 404
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recordUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	record, err := FindRecord(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonDecoder := json.NewDecoder(r.Body)
	err = jsonDecoder.Decode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = record.Update()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recordDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	record := &Record{Id: id}
	err = record.Delete()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
