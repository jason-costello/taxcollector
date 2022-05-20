package scraper

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jason-costello/taxcollector/proxies"
	"github.com/jason-costello/taxcollector/tax"
	"github.com/jason-costello/taxcollector/useragents"
	"golang.org/x/net/publicsuffix"
)

type Scraper struct {
	proxyClient     *proxies.ProxyClient
	db              *sql.DB
	userAgentClient *useragents.UserAgentClient
	httpClient      *http.Client
	currentProxy    proxies.Proxy
	urlsToScrape    []string
}

type Job struct {
	ProcessorID int
	JobID       int
	URL         string
	// DB                 *sql.DB
	// httpClient         *http.Client
	// ProxyClient        *proxies.ProxyClient
	Proxy              proxies.Proxy
	UserAgent          string
	Request            *http.Request
	ResponseBodyBuffer *bytes.Buffer
	PropertyRecord     tax.PropertyRecord
	Duplicate          bool
	Error              error
}

func NewScraper(proxyClient *proxies.ProxyClient, uac *useragents.UserAgentClient, db *sql.DB, httpClient *http.Client, urls []string) *Scraper {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	httpClient.Jar = jar

	return &Scraper{
		httpClient:      httpClient,
		proxyClient:     proxyClient,
		userAgentClient: uac,
		db:              db,
		urlsToScrape:    urls,
	}
}

func (s *Scraper) Scrape() {

	gen := func(done <-chan interface{}, propURL ...string) <-chan Job {
		jobStream := make(chan Job)
		go func() {
			defer close(jobStream)
			for i, u := range propURL {
				job := Job{
					JobID:              i,
					URL:                u,
					Proxy:              proxies.Proxy{},
					UserAgent:          "",
					Request:            nil,
					ResponseBodyBuffer: &bytes.Buffer{},
					PropertyRecord:     tax.PropertyRecord{},
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

					exists, err := s.PropertyExists(j.URL)
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
					j.Proxy, j.Error = s.proxyClient.GetNext()
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
					j.UserAgent, j.Error = s.userAgentClient.GetRandomUserAgent()
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
						resp, j.Error = s.httpClient.Do(firstReq)
						if j.Error == nil {
							j.Error = s.proxyClient.MarkProxyAsBad(j.Proxy.IP)
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

						detailResp, err := s.httpClient.Do(req)
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
					j.Error = s.AddPropertyRecordToDB(j.PropertyRecord)
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
	jobStream := gen(done, s.urlsToScrape...)

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

}
func (s *Scraper) GetURLs() []string {
	return s.urlsToScrape
}
func (s *Scraper) PropertyExists(url string) (bool, error) {
	if s.db == nil {
		return true, errors.New("db is nil")
	}
	urlParts := strings.Split(url, "prop_id=")
	if len(urlParts) < 2 {
		return false, errors.New("no property id provided in url")
	}

	propertyID := strings.TrimSpace(urlParts[1])
	q := "select id from properties where id=? limit 1"
	stmt, err := s.db.Prepare(q)
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

func (s *Scraper) changeUserAgent(req *http.Request) error {
	ua, err := s.userAgentClient.GetRandomUserAgent()
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
	s.httpClient.Transport = transport
	return nil
}

func parseDetails(b *bytes.Buffer) (tax.PropertyRecord, error) {

	doc, err := goquery.NewDocumentFromReader(b)
	if err != nil {
		return tax.PropertyRecord{}, nil
	}
	propertyRecord, err := tax.GetPropertyRecord(doc)
	return propertyRecord, nil
}

func insertLand(pr tax.PropertyRecord, tx *sql.Tx) error {

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

func (s *Scraper) AddPropertyRecordToDB(pr tax.PropertyRecord) error {

	tx, err := s.db.Begin()
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

func insertImprovements(pr tax.PropertyRecord, tx *sql.Tx) error {

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

func insertJurisdictions(pr tax.PropertyRecord, tx *sql.Tx) error {
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
func insertRollValues(pr tax.PropertyRecord, tx *sql.Tx) error {

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
func getRandomTimeoutDuration(min, max int) time.Duration {
	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(max-min) + min

	d, e := time.ParseDuration(fmt.Sprintf("%dms", i))
	if e != nil {
		return time.Millisecond * 1000
	}
	return d
}

func insertPropertyRecord(pr tax.PropertyRecord, tx *sql.Tx) error {
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
