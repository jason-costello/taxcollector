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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jason-costello/taxcollector/proxies"
	"github.com/jason-costello/taxcollector/storage/pgdb"
	"github.com/jason-costello/taxcollector/tax"
	"github.com/jason-costello/taxcollector/useragents"
	"golang.org/x/net/publicsuffix"
)

// TODO:  Add flags to pull in db creds, worker count,
type Scraper struct {
	proxyClient     *proxies.ProxyClient
	db              *sql.DB
	pdb             *pgdb.Queries
	userAgentClient *useragents.UserAgentClient
	httpClient      *http.Client
	currentProxy    proxies.Proxy
}

func NewScraper(proxyClient *proxies.ProxyClient, uac *useragents.UserAgentClient, db *sql.DB, httpClient *http.Client) *Scraper {
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
		pdb:             pgdb.New(db),
	}
}

func (s *Scraper) Scrape(urls []string) {
	var workers = runtime.NumCPU()

	jobChannel := make(chan Job)

	go func() {
		for i, u := range urls {

			var propID = ""
			parts := strings.Split(u, "=")
			if len(parts) > 1 {
				propID = parts[2]
			}
			if propID == "" {
				continue
			}
			jobChannel <- Job{
				ProcessorID:        0,
				JobID:              i,
				URL:                u,
				Proxy:              proxies.Proxy{},
				UserAgent:          "",
				Request:            nil,
				ResponseBodyBuffer: nil,
				PropertyRecord:     tax.PropertyRecord{PropertyID: propID},
				Duplicate:          false,
				Error:              nil,
				Scraper:            s,
			}
		}
	}()
	jobResultsChan := make(chan Job)
	wg := &sync.WaitGroup{}
	wg.Add(workers)
	go func() {
		wg.Wait()
		close(jobResultsChan)
	}()

	for i := 1; i <= workers; i++ {
		go func(id int) {
			defer wg.Done()

			for j := range jobChannel {
				j.ProcessorID = id
				j.Process()
				jobResultsChan <- j
				time.Sleep(time.Second)

			}
		}(i)
	}

	for r := range jobResultsChan {
		if r.Error == nil {
			r.Error = errors.New("No Error")
		}
		fmt.Printf("worker: %d   job: %d propertyID: %s  final error: %s\n", r.ProcessorID, r.JobID, r.PropertyRecord.PropertyID, r.Error)
	}

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
	pid, err := strconv.Atoi(propertyID)
	if err != nil {
		return false, err
	}

	if pid == 0 {
		return true, errors.New("invalid property id: 0")
	}
	prop, err := s.pdb.GetPropertyByID(context.Background(), int32(pid))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}
	if prop.Address.Valid && prop.Address.String != "" {
		return true, nil
	}
	return false, nil
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
func stringToNullInt32(s string) sql.NullInt32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		i = 0
	}
	return sql.NullInt32{
		Int32: int32(i),
		Valid: true,
	}
}

func stringToFloat64(s string) sql.NullFloat64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		i = 0.0
	}
	return sql.NullFloat64{
		Float64: i,
		Valid:   true,
	}
}

func stringToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
func insertLand(pdb *pgdb.Queries, pr tax.PropertyRecord, tx *sql.Tx) error {

	for _, i := range pr.Land {

		landParams := pgdb.InsertLandParams{
			Number:      stringToNullInt32(i.Number),
			LandType:    stringToNullString(i.Type),
			Description: stringToNullString(i.Description),
			Acres:       stringToFloat64(i.Acres),
			SquareFeet:  stringToFloat64(i.Sqft),
			EffFront:    stringToFloat64(i.EffFront),
			EffDepth:    stringToFloat64(i.EffDepth),
			MarketValue: stringToNullInt32(i.MarketValue),
			PropertyID:  stringToNullInt32(pr.PropertyID),
		}
		if err := pdb.WithTx(tx).InsertLand(context.Background(), landParams); err != nil {
			tx.Rollback()
			return err
		}

	}

	return nil
}

func (s *Scraper) AddPropertyRecordToDB(workerID, jobID int, pUrl string, pr tax.PropertyRecord) error {

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("worker: %d  job: %d propID: %s - error s.db.Begin() error: %w\n", workerID, jobID, pr.PropertyID, err)
	}

	err = insertPropertyRecord(s.pdb, pr, tx)
	if err != nil {
		return fmt.Errorf("worker: %d  job: %d propID: %s - insertPropertyRecord error: %w\n", workerID, jobID, pr.PropertyID, err)
		tx.Rollback()

		return err
	}

	err = insertRollValues(s.pdb, pr, tx)
	if err != nil {
		return fmt.Errorf("worker: %d  job: %d  propID: %s - Status: insertRollValues error: %w\n", workerID, jobID, pr.PropertyID, err)
		tx.Rollback()
		return err
	}

	err = insertJurisdictions(s.pdb, pr, tx)
	if err != nil {
		return fmt.Errorf("worker: %d  job: %d propID: %s - Status: insertJurisdictions error: %w\n", workerID, jobID, pr.PropertyID, err)
		tx.Rollback()
		return err
	}

	err = insertImprovements(s.pdb, pr, tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("worker: %d  job: %d  propID: %s - Status: insertImprovements error: %w\n", workerID, jobID, pr.PropertyID, err)
	}

	err = insertLand(s.pdb, pr, tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("worker: %d  job: %d propID: %s - Status: insertLand error: %w\n", workerID, jobID, pr.PropertyID, err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("worker: %d  job: %d  propID: %s - Error on tx.Commit: %w\n", workerID, jobID, pr.PropertyID, err)
	}
	fmt.Printf("worker: %d  job: %d  propID: %s - All records committed\n", workerID, jobID, pr.PropertyID)

	return nil
}

func insertImprovements(pdb *pgdb.Queries, pr tax.PropertyRecord, tx *sql.Tx) error {

	for _, i := range pr.Improvements {
		params := pgdb.InsertImprovementParams{
			Name:        stringToNullString(i.Name),
			Description: stringToNullString(i.Description),
			StateCode:   stringToNullString(i.StateCode),
			LivingArea:  stringToNullInt32(i.LivingArea),
			Value:       stringToNullInt32(i.Value),
			PropertyID:  stringToNullInt32(pr.PropertyID),
		}

		id, err := pdb.WithTx(tx).InsertImprovement(context.Background(), params)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, d := range i.Details {
			paramDetails := pgdb.InsertImprovementDetailParams{
				ImprovementID:   sql.NullInt32{Int32: id, Valid: true},
				ImprovementType: stringToNullString(d.Type),
				Description:     stringToNullString(d.Description),
				Class:           stringToNullString(d.Class),
				ExteriorWall:    stringToNullString(d.ExteriorWall),
				YearBuilt:       stringToNullInt32(d.YearBuilt),
				SquareFeet:      stringToNullInt32(d.SqFt),
			}

			if err := pdb.WithTx(tx).InsertImprovementDetail(context.Background(), paramDetails); err != nil {
				tx.Rollback()
				return err
			}

		}
	}
	return nil
}

func insertJurisdictions(pdb *pgdb.Queries, pr tax.PropertyRecord, tx *sql.Tx) error {
	for _, j := range pr.Jurisdictions {

		params := pgdb.InsertJurisdictionParams{
			Entity:         sql.NullString{},
			Description:    sql.NullString{},
			TaxRate:        stringToNullInt32(j.TaxRate),
			AppraisedValue: stringToNullInt32(j.AppraisedValue),
			TaxableValue:   stringToNullInt32(j.TaxableValue),
			EstimatedTax:   stringToNullInt32(j.EstimatedTax),
			PropertyID:     stringToNullInt32(pr.PropertyID),
		}

		if err := pdb.WithTx(tx).InsertJurisdiction(context.Background(), params); err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil

}
func insertRollValues(pdb *pgdb.Queries, pr tax.PropertyRecord, tx *sql.Tx) error {

	for _, r := range pr.RollValue {

		rollParams := pgdb.InsertRollValueParams{
			Year:         stringToNullInt32(r.Year),
			Improvements: stringToNullInt32(r.Improvements),
			LandMarket:   stringToNullInt32(r.LandMarket),
			AgValuation:  stringToNullInt32(r.AgValuation),
			Appraised:    stringToNullInt32(r.Appraised),
			HomesteadCap: stringToNullInt32(r.HomesteadCap),
			Assessed:     stringToNullInt32(r.Assessed),
			PropertyID:   stringToNullInt32(pr.PropertyID),
		}

		if err := pdb.WithTx(tx).InsertRollValue(context.Background(), rollParams); err != nil {
			tx.Rollback()
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

func stringToInt32(s string) int32 {

	i, e := strconv.Atoi(s)
	if e != nil {
		return int32(0)
	}
	return int32(i)

}
func insertPropertyRecord(pdb *pgdb.Queries, pr tax.PropertyRecord, tx *sql.Tx) error {

	propParams := pgdb.InsertPropertyRecordParams{
		ID:                  stringToInt32(pr.PropertyID),
		OwnerID:             stringToNullInt32(pr.OwnerID),
		OwnerName:           stringToNullString(pr.OwnerName),
		OwnerMailingAddress: stringToNullString(pr.OwnerMailingAddress),
		Zoning:              stringToNullString(pr.Zoning),
		NeighborhoodCd:      stringToNullString(pr.NeighborhoodCD),
		Neighborhood:        stringToNullString(pr.Neighborhood),
		Address:             stringToNullString(pr.Address),
		LegalDescription:    stringToNullString(pr.LegalDescription),
		GeographicID:        stringToNullString(pr.GeographicID),
		Exemptions:          stringToNullString(pr.Exemptions),
		OwnershipPercentage: stringToFloat64(pr.OwnershipPercentage),
		MapscoMapID:         stringToNullString(pr.MapscoMapID),
	}
	if err := pdb.WithTx(tx).InsertPropertyRecord(context.Background(), propParams); err != nil {
		fmt.Printf("propID: %s     Err property insert:  %s\n", pr.PropertyID, err)
		fmt.Printf("%#+v\n", propParams)
		tx.Rollback()
		return err
	}

	return nil

}

type Job struct {
	ProcessorID        int
	JobID              int
	URL                string
	Proxy              proxies.Proxy
	UserAgent          string
	Request            *http.Request
	ResponseBodyBuffer *bytes.Buffer
	PropertyRecord     tax.PropertyRecord
	Duplicate          bool
	Error              error
	Scraper            *Scraper
}

func (j *Job) ProcessError(removeURL bool, fun string, nerr error) error {
	if removeURL {
		if err := j.Scraper.pdb.RemovePendingURL(context.Background(), j.URL); err != nil {
			return err
		}
	}
	fmt.Printf("worker: %d   job: %d   propertyID: %s  function: %s  error during processing: %s\n", j.ProcessorID, j.JobID, j.PropertyRecord.PropertyID, fun, nerr)
	return nil
}
func (j *Job) Process() {

	var propID int
	var property pgdb.Property

	if j.PropertyRecord.PropertyID == "" {
		j.ProcessError(false, "strconv.Atoi(j.PropertyRecord.PropertyID)", errors.New("no property record id set"))
		return

	}
	propID, j.Error = strconv.Atoi(j.PropertyRecord.PropertyID)
	if j.Error != nil {
		j.ProcessError(false, "strconv.Atoi(j.PropertyRecord.PropertyID)", j.Error)
		return
	}

	property, j.Error = j.Scraper.pdb.GetPropertyByID(context.Background(), int32(propID))
	if j.Error != nil {
		if j.Error.Error() != "sql: no rows in result set" {
			j.ProcessError(true, "GetPropertyByID", j.Error)
			return
		}
	}

	if property.ID == int32(propID) {
		j.Error = errors.New("duplicate ID")
		j.ProcessError(true, fmt.Sprintf("propertyID: %d == propID: %d ", property.ID, propID), j.Error)
		return
	}

	j.Proxy, j.Error = j.Scraper.proxyClient.GetNext()
	if j.Error != nil {
		j.ProcessError(false, "proxyClient.GetNext()", j.Error)
		return
	}

	fmt.Printf("worker: %d   jobID: %d propID: %s   Getting user agent\n", j.ProcessorID, j.JobID, j.PropertyRecord.PropertyID)
	j.UserAgent, j.Error = j.Scraper.userAgentClient.GetRandomUserAgent()
	if j.Error != nil {
		j.ProcessError(false, "userAgentClient.GetRandomUserAgent()", j.Error)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var firstReq *http.Request
	firstReq, j.Error = http.NewRequestWithContext(ctx, "GET", "https://propaccess.trueautomation.com/clientdb/?cid=56", nil)
	if j.Error != nil {
		j.ProcessError(false, "http.NewRequestWithContext", j.Error)
		return
	}
	var resp *http.Response
	resp, j.Error = j.Scraper.httpClient.Do(firstReq)
	if j.Error != nil {
		fmt.Printf("worker: %d   jobID: %d  Bad proxy\n", j.ProcessorID, j.JobID)
		j.Error = j.Scraper.proxyClient.MarkProxyAsBad(j.Proxy.IP)
		dur := getRandomTimeoutDuration(10, 100)
		time.Sleep(dur)
	}
	if j.Error != nil {
		j.ProcessError(false, "http.proxyClient.MarkProxyAsBad", j.Error)
		return
	}
	if resp.StatusCode > 399 || resp.StatusCode < 200 {
		j.Error = errors.New(resp.Status)
		j.ProcessError(false, "http.proxyClient.MarkProxyAsBad", j.Error)
		return

	}

	var req *http.Request
	req, j.Error = http.NewRequest("GET", j.URL, nil)
	if j.Error != nil {
		j.ProcessError(false, "http.NewRequest", j.Error)
		return
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Host", "propaccesj.Scraper.trueautomation.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://propaccesj.Scraper.trueautomation.com/clientdb/SearchResultj.Scraper.aspx?cid=56")
	fmt.Printf("worker: %d   jobID: %d  Property Request\n", j.ProcessorID, j.JobID)

	var detailResp *http.Response
	detailResp, j.Error = j.Scraper.httpClient.Do(req)

	if j.Error != nil {
		j.ProcessError(false, "j.Scraper.httpClient.Do", j.Error)
		return
	}

	var b []byte
	b, j.Error = io.ReadAll(detailResp.Body)
	if j.Error != nil {
		j.ProcessError(false, "io.ReadAll(detailResp.Body)", j.Error)
		return
	}
	j.ResponseBodyBuffer = bytes.NewBuffer(b)

	detailResp.Body.Close()

	if j.ResponseBodyBuffer == nil {
		j.Error = errors.New("nil response body")
		j.ProcessError(false, "j.ResponseBodyBuffer == nil", j.Error)
		return
	}

	fmt.Printf("worker: %d   jobID: %d  parsing property details\n", j.ProcessorID, j.JobID)
	j.PropertyRecord, j.Error = parseDetails(j.ResponseBodyBuffer)
	if j.Error != nil {
		j.ProcessError(false, "parseDetails(j.ResponseBodyBuffer)", j.Error)
		return
	}

	fmt.Printf("worker: %d   jobID: %d  adding records to database\n", j.ProcessorID, j.JobID)

	if j.Error = j.Scraper.AddPropertyRecordToDB(j.ProcessorID, j.JobID, j.URL, j.PropertyRecord); j.Error != nil {
		j.ProcessError(false, "j.Scraper.AddPropertyRecordToDB()", j.Error)
		return
	}

	if j.Error = j.Scraper.pdb.RemovePendingURL(context.Background(), j.URL); j.Error != nil {
		j.ProcessError(false, "j.Scraper.pdb.RemovePendingUR", j.Error)
		return
	}

	if j.Error == nil {
		j.Error = errors.New("No Errors")
	}
	fmt.Printf("worker: %d  jobID: %d  procID:  %d   finalErrors: %s\n", j.ProcessorID, j.JobID, j.ProcessorID, j.Error)
}
