package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jason-costello/taxcollector/storage/pgdb"
)

type Handler struct {
	db *pgdb.Queries
}

func NewHandler(db *pgdb.Queries) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(200)
	w.Write([]byte("v0.0.1"))

}

func (h *Handler) GetProperty(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	if vars == nil {
		w.WriteHeader(500)
		w.Write([]byte("no propertyID provided"))
		return
	}
	propertyID := vars["id"]
	pid, err := strconv.Atoi(propertyID)
	if err != nil {

	}
	h.db.GetPropertyByID()
}
