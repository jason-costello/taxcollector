package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jason-costello/taxcollector/storage/pgdb"
	_ "github.com/lib/pq"

	"github.com/jason-costello/taxcollector/proxies"
	"github.com/jason-costello/taxcollector/scraper"
	"github.com/jason-costello/taxcollector/useragents"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"192.168.1.100", 5432, "postgres", "postgres", "tax")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	uac := &useragents.UserAgentClient{}
	uac.LoadUserAgents("../../useragents.txt")
	pc := proxies.NewProxyClient(db)

	if err != nil {
		panic(err)
	}
	s := scraper.NewScraper(pc, uac, db, nil)
	pdb := pgdb.New(db)

	remainingCount, err := pdb.GetRemainingURLCount(context.Background())
	if err != nil {
		panic(err)
	}

	urls, err := pdb.GetRandomURLs(context.Background(), int32(remainingCount))
	if err != nil {
		panic(err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(len(urls))
	s.Scrape(urls)
	wg.Wait()
}
