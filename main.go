package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/publicsuffix"
)

type Job struct {
	ProcessorID        int
	JobID              int
	URL                string
	DB                 *sql.DB
	httpClient         *http.Client
	ProxyClient        *ProxyClient
	Proxy              Proxy
	UserAgent          string
	Request            *http.Request
	ResponseBodyBuffer *bytes.Buffer
	PropertyRecord     PropertyRecord
	Duplicate          bool
	Error              error
}

func propertyExists(url string, db *sql.DB) (bool, error) {
	if db == nil {
		return true, errors.New("db is nil")
	}
	urlParts := strings.Split(url, "prop_id=")
	if len(urlParts) < 2 {
		return false, errors.New("no property id provided in url")
	}

	propertyID := strings.TrimSpace(urlParts[1])
	q := "select id from properties where id=? limit 1"
	stmt, err := db.Prepare(q)
	if err != nil {
		return false, err
	}
	q = strings.Replace(q, "?", urlParts[1], -1)
	// row := stmt.QueryRow(urlParts[1])
	row := stmt.QueryRow(propertyID)

	var pID string
	row.Scan(&pID)
	if pID == "" {
		return false, nil
	}
	return true, nil
}
func main() {

	go func() {
		if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
			log.Fatal(err)
		}
	}()

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pc := NewProxyClient(db)
	// userAgents, err := loadUserAgents("useragents.txt")
	// if err != nil {
	// 	panic(err)
	// }

	propertyURLS, err := readLines("./urls-to-fetch.txt")
	if err != nil {
		panic(err)
	}
	userAgents, err := loadUserAgents("useragents.txt")
	if err != nil {
		panic(err)
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: jar,
	}
	scrape(propertyURLS, pc, client, userAgents, db)
}

func scrape(urls []string, proxyClient *ProxyClient, hc *http.Client, userAgents []string, db *sql.DB) {

	gen := func(done <-chan interface{}, propURL ...string) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for i, u := range propURL {
				job := Job{
					JobID:              i,
					URL:                u,
					DB:                 db,
					httpClient:         hc,
					ProxyClient:        proxyClient,
					Proxy:              Proxy{},
					UserAgent:          "",
					Request:            nil,
					ResponseBodyBuffer: &bytes.Buffer{},
					PropertyRecord:     PropertyRecord{},
					Error:              nil,
					Duplicate:          false,
				}
				select {
				case <-done:
					return
				case jobStream <- job:
				}
			}
		}()
		return jobStream
	}

	checkIfDuplicate := func(done <-chan interface{}, processorID int, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)

		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				j.ProcessorID = processorID
				if j.Error == nil {

					exists, err := propertyExists(j.URL, j.DB)
					if err != nil {
						j.Error = err

					} else if exists {
						j.Duplicate = true
						j.Error = errors.New("duplicate record")
					}
				}
				select {
				case <-done:
					return
				case jobStream <- j:

				}
			}
		}()

		return jobStream
	}
	getProxy := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {

		jobStream := make(chan Job)

		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				if j.Error == nil {
					j.Proxy, j.Error = j.ProxyClient.GetNext()
				}

				select {
				case <-done:
					return
				case jobStream <- j:
				}
			}

		}()
		return jobStream
	}
	getUserAgent := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)

			for j := range incomingJobStream {
				if j.Error == nil {
					j.UserAgent, j.Error = getRandomUserAgent(userAgents)
				}
				select {
				case <-done:
					return
				case jobStream <- j:

				}
			}
		}()
		return jobStream
	}

	buildRequest := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				if j.Error == nil {

				}
				select {
				case <-done:
					return
				case jobStream <- j:

				}
			}
		}()
		return jobStream
	}
	pingPage := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if j.Error == nil {
					var firstReq *http.Request
					firstReq, j.Error = http.NewRequestWithContext(ctx, "GET", "https://propaccess.trueautomation.com/clientdb/?cid=56", nil)

					if j.Error == nil {

						var resp *http.Response
						resp, j.Error = j.httpClient.Do(firstReq)
						if j.Error == nil {
							j.Error = markProxyAsBad(j.DB, j.Proxy.IP)
							dur := getRandomTimeoutDuration(10, 250)
							time.Sleep(dur)
						}
						if j.Error == nil {
							if resp.StatusCode > 399 || resp.StatusCode < 200 {
								j.Error = errors.New(resp.Status)
							}
						}

					}
					select {
					case <-done:
						return
					case jobStream <- j:

					}
				}
			}
		}()
		return jobStream
	}
	propertyRequest := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				if j.Error == nil {
					var req *http.Request
					req, j.Error = http.NewRequest("GET", j.URL, nil)
					if j.Error == nil {
						req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
						req.Header.Set("Accept-Encoding", "gzip, deflate, br")
						req.Header.Set("Host", "propaccess.trueautomation.com")
						req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15")
						req.Header.Set("Accept-Language", "en-US,en;q=0.9")
						req.Header.Set("Referer", "https://propaccess.trueautomation.com/clientdb/SearchResults.aspx?cid=56")

						detailResp, err := j.httpClient.Do(req)
						if err != nil {
							j.Error = err
						} else {

							b, err := io.ReadAll(detailResp.Body)
							if err != nil {
								j.Error = err
							} else {
								j.ResponseBodyBuffer = bytes.NewBuffer(b)
							}

							detailResp.Body.Close()
						}
					}
					select {
					case <-done:
						return
					case jobStream <- j:

					}
				}
			}

		}()
		return jobStream
	}

	parseResponse := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				if j.Error == nil {

					if j.ResponseBodyBuffer != nil {
						j.PropertyRecord, j.Error = parseDetails(j.ResponseBodyBuffer)
					}
					select {
					case <-done:
						return
					case jobStream <- j:

					}
				}
			}
		}()
		return jobStream
	}

	writeToDB := func(done <-chan interface{}, incomingJobStream <-chan Job) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for j := range incomingJobStream {
				if j.Error == nil {
					j.Error = AddPropertyRecordToDB(j.PropertyRecord, db)
				}
				select {
				case <-done:
					return
				case jobStream <- j:

				}
			}
		}()
		return jobStream
	}

	fanIn := func(done <-chan interface{}, channels ...<-chan Job) <-chan Job {
		var wg sync.WaitGroup
		mplexStream := make(chan Job)

		mplex := func(c <-chan Job) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case mplexStream <- i:
				}
			}
		}
		wg.Add(len(channels))
		for _, c := range channels {
			go mplex(c)
		}

		go func() {
			wg.Wait()
			close(mplexStream)
		}()

		return mplexStream
	} // end fan-in

	done := make(chan interface{})
	defer close(done)
	jobStream := gen(done, urls...)

	procCount := 7
	processors := make([]<-chan Job, procCount)

	for i := 0; i < procCount; i++ {
		processors[i] = writeToDB(done, parseResponse(done, propertyRequest(done, pingPage(done, buildRequest(done, getUserAgent(done, getProxy(done, checkIfDuplicate(done, i, jobStream))))))))
	}

	var results []Job
	// fanIn used to consolidate all the results to rs
	for r := range fanIn(done, processors...) {
		if r.Error == nil {
			fmt.Println("INFO : ", r.ProcessorID, r.JobID, r.URL)
		} else {
			fmt.Println("ERROR: ", r.ProcessorID, r.JobID, r.URL, r.Error)
		}
		results = append(results, r)

	}

	// for _, r := range results {
	// 	if r.Error == nil {
	// 		fmt.Println("INFO : ", r.ProcessorID, r.JobID, r.URL)
	// 	} else {
	// 		fmt.Println("ERROR: ", r.ProcessorID, r.JobID, r.URL, r.Error)
	// 	}
	// }
}

// / old stuff below

// scrape(propertyURLS, proxyClient)

// scraper := NewScraper(pc, userAgents, db)

// propertyURLS := []string{
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=2290",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=32543",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=411413",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=71352",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=44239",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=401253",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=114173",
// 	"https://propaccess.trueautomation.com/clientdb/Property.aspx?cid=56&prop_id=443560",
// }

// 	for _, propertyURL := range propertyURLS {
//
// 		recordExists, err := propertyExists(propertyURL, db)
// 		if recordExists {
// 			fmt.Println("property already in database")
// 			continue
// 		}
//
// 		if err != nil {
// 			fmt.Println("error checking db for duplicate property, attempt to collect")
// 		}
//
// 		dur := getRandomTimeoutDuration(150, 836)
// 		time.Sleep(dur)
// 		var pageDetails []byte
// 		if pageDetails, err = scraper.Scrape(propertyURL); err != nil {
// 			fmt.Println("ERROR: ", err.Error())
// 			continue
// 		}
// 		propertyRecord, err := parseDetails(pageDetails)
// 		if err != nil {
// 			fmt.Println("ERROR: ", err)
// 			continue
// 		}
//
// 		if err := AddPropertyRecordToDB(propertyRecord, db); err != nil {
// 			fmt.Println("ERROR: ", err)
// 			continue
// 		}
//
// 		// prBytes, _ := json.Marshal(propertyRecord)
// 		// if err := bytesToFile(propertyURL, prBytes); err != nil {
// 		// 	fmt.Println("ERROR: ", err)
// 		// 	continue
// 		// }
//
// 	}
//
// 	fmt.Println("DONE")
// }
//

// func bytesToFile(propertyURL string, pageDetails []byte) error {
// 	err := os.WriteFile(fmt.Sprintf("%s.json", strings.Split(propertyURL, "prop_id=")[1]), pageDetails, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

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

func loadUserAgents(filename string) ([]string, error) {

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var agents []string
	for _, line := range strings.Split(string(b), "\n") {
		if line != "" {
			agents = append(agents, line)
		}
	}
	return agents, nil

}
func getRandomUserAgent(agentList []string) (string, error) {

	if agentList == nil {
		return "", errors.New("agent list is nil")
	}
	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(len(agentList)-0) + 0

	return agentList[i], nil

}
func getRandomTimeoutDuration(min, max int) time.Duration {
	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(max-min) + min

	d, e := time.ParseDuration(fmt.Sprintf("%dms", i))
	if e != nil {
		return time.Millisecond * 1000
	}
	return d
}

func markProxyAsBad(db *sql.DB, proxyIP string) error {
	updateQuery := `update proxies set isBad = true where ip = ?`
	updateStmt, err := db.Prepare(updateQuery)
	if err != nil {
		return err
	}

	_, err = updateStmt.Exec(proxyIP)
	if err != nil {
		return err
	}
	return nil
}

//
// func (s *Scraper) Scrape(url string) ([]byte, error) {
// 	pidParts := strings.Split(url, "prop_id=")
// 	if len(pidParts) < 2 {
// 		return nil, errors.New("no property id found in url")
// 	}
// 	pid := pidParts[1]
// 	fmt.Printf("\n%s: ", pid)
// 	s.addProxy()
// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
// 	defer cancel()
//
// 	firstReq, err := http.NewRequestWithContext(ctx, "GET", "https://propaccess.trueautomation.com/clientdb/?cid=56", nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if err := s.changeUserAgent(firstReq); err != nil {
// 		return nil, err
// 	}
// 	resp, err := s.hc.Do(firstReq)
// 	if err != nil {
// 		if err := markProxyAsBad(); err != nil {
// 			fmt.Println("Unable to mark proxy as bad: ", s.currentProxy.IP)
// 		}
// 		fmt.Println("bad proxy")
// 		dur := getRandomTimeoutDuration(125, 500)
// 		time.Sleep(dur)
// 		return s.Scrape(url)
// 	}
// 	if resp.StatusCode > 399 || resp.StatusCode < 200 {
// 		return nil, errors.New(resp.Status)
// 	}
//
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
// 	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
// 	req.Header.Set("Host", "propaccess.trueautomation.com")
// 	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15")
// 	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
// 	req.Header.Set("Referer", "https://propaccess.trueautomation.com/clientdb/SearchResults.aspx?cid=56")
//
// 	detailResp, err := s.hc.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	fmt.Println(detailResp.Status)
// 	if detailResp.StatusCode > 399 || detailResp.StatusCode < 200 {
// 		return nil, errors.New(detailResp.Status)
// 	}
//
// 	defer detailResp.Body.Close()
//
// 	b, err := ioutil.ReadAll(detailResp.Body)
// 	if err != nil {
// 		panic(b)
// 	}
//
// 	return b, nil
// }

type Scraper struct {
	hc           *http.Client
	proxyClient  *ProxyClient
	userAgents   []string
	DB           *sql.DB
	currentProxy Proxy
}

func NewScraper(proxyClient *ProxyClient, userAgents []string, db *sql.DB) *Scraper {
	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	return &Scraper{
		hc:          client,
		proxyClient: proxyClient,
		userAgents:  userAgents,
		DB:          db,
	}
}
func (s *Scraper) changeUserAgent(req *http.Request) error {
	ua, err := getRandomUserAgent(s.userAgents)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", ua)
	return nil
}
func (s *Scraper) addProxy() error {
	p, err := s.proxyClient.GetNext()
	if err != nil {
		return err
	}
	proxyStr := fmt.Sprintf("http://%s", p.IP)

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	s.currentProxy = p
	s.hc.Transport = transport
	return nil
}

func parseDetails(b *bytes.Buffer) (PropertyRecord, error) {

	doc, err := goquery.NewDocumentFromReader(b)
	if err != nil {
		return PropertyRecord{}, nil
	}
	propertyRecord, err := getPropertyRecord(doc)
	return propertyRecord, nil
}
func insertLand(pr PropertyRecord, tx *sql.Tx) error {

	for _, i := range pr.Land {

		query := "insert into land(number, type, description, acres, squareFeet, effFront, effDepth, marketValue, propertyID) values(?,?,?,?,?,?,?,?,?)"
		stmt, err := tx.Prepare(query)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(i.Number, i.Type, i.Description, i.Acres, i.Sqft, i.EffFront, i.EffDepth, i.MarketValue, pr.PropertyID); err != nil {
			return err
		}
	}

	return nil
}

func insertPropertyRecord(pr PropertyRecord, tx *sql.Tx) error {
	query := `insert into properties(id,ownerID,ownerName,ownerMailingAddress,
										zoning,neighborhoodCD,neighborhood,
										address, legalDescription, geographicID, exemptions,
										ownershipPercentage, mapscoMapID) 
			values(?,?,?,?,?,?,?,?,?,?,?,?,?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(pr.PropertyID, pr.OwnerID, pr.OwnerName, pr.OwnerMailingAddress,
		pr.Zoning, pr.NeighborhoodCD, pr.Neighborhood,
		pr.Address, pr.LegalDescription, pr.GeographicID, pr.Exemptions,
		pr.OwnershipPercentage, pr.MapscoMapID)

	if err != nil {
		return err
	}

	return nil

}

func AddPropertyRecordToDB(pr PropertyRecord, db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {

		return err
	}

	err = insertPropertyRecord(pr, tx)
	if err != nil {
		tx.Rollback()

		return err
	}

	err = insertRollValues(pr, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertJurisdictions(pr, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertImprovements(pr, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = insertLand(pr, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func insertImprovements(pr PropertyRecord, tx *sql.Tx) error {

	for _, i := range pr.Improvements {

		iQuery := "insert into improvements (name, description, stateCode, livingArea, value, propertyID) values(?,?,?,?,?,?)"
		stmt, err := tx.Prepare(iQuery)
		if err != nil {
			return err
		}
		result, err := stmt.Exec(i.Name, i.Description, i.StateCode, i.LivingArea, i.Value, pr.PropertyID)
		if err != nil {
			return err
		}
		improveID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		for _, d := range i.Details {

			dQuery := "insert into improvementDetail(improvementID, type, description, class, exteriorWall, yearBuilt, squareFeet) values (?,?,?,?,?,?,?)"
			dStmt, err := tx.Prepare(dQuery)
			if err != nil {
				return err
			}
			if _, err := dStmt.Exec(improveID, d.Type, d.Description, d.Class, d.ExteriorWall, d.YearBuilt, d.SqFt); err != nil {
				return err
			}

		}

	}
	return nil
}

func insertJurisdictions(pr PropertyRecord, tx *sql.Tx) error {
	for _, j := range pr.Jurisdictions {

		query := "insert into jurisdictions( entity, description, taxRate, appraisedValue, taxableValue, estimatedTax, propertyID) values(?,?,?,?,?,?,?)"

		stmt, err := tx.Prepare(query)
		if err != nil {
			fmt.Println("error inserting jurisdictions: ", err)
			return err
		}
		if _, err := stmt.Exec(j.Entity, j.Description, j.TaxRate, j.AppraisedValue, j.TaxableValue, j.EstimatedTax, pr.PropertyID); err != nil {
			return err
		}

	}
	return nil

}
func insertRollValues(pr PropertyRecord, tx *sql.Tx) error {

	for _, r := range pr.RollValue {

		query := "insert into rollValues( year, improvements, landMarket, agValuation, appraised, homesteadCap, assessed, propertyID) values(?,?,?,?,?,?,?,?)"

		stmt, err := tx.Prepare(query)
		if err != nil {
			fmt.Println("error inserting roll values: ", err)

			return err
		}

		if _, err := stmt.Exec(r.Year, r.Improvements, r.LandMarket, r.AgValuation, r.Appraised, r.HomesteadCap, r.Assessed, pr.PropertyID); err != nil {
			return err
		}

	}
	return nil
}

// func CreateProxyTable(db *sql.DB) error {
//
// 	sqlStmt := `create table if not exists  proxies (ip varchar(30) not null primary key, lastused datetime);`
// 	_, err := db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlStmt)
// 		return err
// 	}
// 	return nil
// }

// func LoadProxyTable(db *sql.DB, proxies []*Proxy) error {
//
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}
//
// 	stmt, err := tx.Prepare("insert into proxies(ip, lastUsed) values(?, ?)")
// 	if err != nil {
// 		return err
// 	}
//
// 	defer stmt.Close()
// 	for _, proxy := range proxies {
// 		_, err = stmt.Exec(proxy.IP, proxy.LastUsed)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	tx.Commit()
// 	return nil
//
// }
