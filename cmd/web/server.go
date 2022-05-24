package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jason-costello/taxcollector/web"
	_ "github.com/lib/pq"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"127.0.0.1", 5432, "postgres", "password", "tax")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	hs := http.Server{}
	server := web.NewServer(db, &hs)

	server.Serve()

}
