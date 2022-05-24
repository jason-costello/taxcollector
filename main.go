package main

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/jason-costello/taxcollector/web"
	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/jason-costello/taxcollector/proxies"
	"github.com/jason-costello/taxcollector/scraper"
	"github.com/jason-costello/taxcollector/storage/pgdb"
	"github.com/jason-costello/taxcollector/useragents"
	_ "github.com/mattn/go-sqlite3"
)

var taxDB *pgdb.Queries

func main() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}

	taxDB = pgdb.New(db)
	defer db.Close()

	handler := web.NewHandler(taxDB)

	r := mux.NewRouter()
	v1ApiRouter := r.PathPrefix("/v1/api").Subrouter()
	v1ApiRouter.HandleFunc("/property/{id}", handler.GetProperty)

	r.HandleFunc("/version", handler.Version)

	http.ListenAndServe(":8888", r)
}

func getPropertyByID(c *gin.Context) {
	id := c.Param("id")

	i, err := strconv.Atoi(id)
	if err != nil {
		c.Error(err)
		c.Status(407)
	}

	property, err := taxDB.GetPropertyByID(context.Background(), int32(i))
	if err != nil {
		c.Error(err)
		c.Status(407)

	}
	c.JSON(http.StatusOK, gin.H{"data": property})
}
func loadScraper() (*scraper.Scraper, error) {

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pc := proxies.NewProxyClient(db)

	ua := useragents.UserAgentClient{}
	if err := ua.LoadUserAgents("useragents.txt"); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	hc := http.DefaultClient

	scraper := scraper.NewScraper(pc, &ua, db, hc)

	return scraper, nil

}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
