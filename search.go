package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/net/publicsuffix"
)

type Client struct {
	*http.Client
}

func NewClient(hc *http.Client) *Client {

	if hc == nil {
		hc = &http.Client{}
	}

	if hc.Jar == nil {
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			log.Fatal(err)
		}
		hc.Jar = jar
	}
	return &Client{hc}
}

func (c *Client) GrabSession(urlStr string) error {

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode > 399 || resp.StatusCode < 200 {
		return errors.New(resp.Status)
	}

	valMap, err := getHiddenValues(resp)
	if err != nil {
		return err
	}

	PrintMap(valMap)

	payload := strings.NewReader(fmt.Sprintf(`__EVENTTARGET=%s&__EVENTARGUMENT=%s&__VIEWSTATE=%s&__VIEWSTATEGENERATOR=%s&__EVENTVALIDATION=%s&propertySearchOptions%%3AsearchText=Magazine&propertySearchOptions%%3Asearch=Search&propertySearchOptions%%3AownerName=&propertySearchOptions%%3AstreetNumber=&propertySearchOptions%%3AstreetName=&propertySearchOptions%%3Apropertyid=&propertySearchOptions%%3Ageoid=&propertySearchOptions%%3Adba=&propertySearchOptions%%3Aabstract=&propertySearchOptions%%3Asubdivision=&propertySearchOptions%%3AmobileHome=&propertySearchOptions%%3Acondo=&propertySearchOptions%%3AagentCode=&propertySearchOptions%%3Ataxyear=2022&propertySearchOptions%%3ApropertyType=All&propertySearchOptions%%3AorderResultsBy=Owner+Name&propertySearchOptions%%3ArecordsPerPage=250`, valMap["__EVENTTARGET"], valMap["__EVENTARGUMENT"], valMap["__VIEWSTATE"], valMap["__VIEWSTATEGENERATOR"], valMap["__EVENTVALIDATION"]))

	req, err = http.NewRequest("POST",
		"https://propaccess.trueautomation.com/ClientDB/SearchResults.aspx?cid=56", payload)
	if err != nil {
		return err
	}

	// req.Header.Add("Cookie", "ASP.NET_SessionId=rh4xkq45w3cgearhtxmdgx45")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Host", "propaccess.trueautomation.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Referer", "https://propaccess.trueautomation.com/ClientDB/PropertySearch.aspx?cid=56")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "text/plain; charset=utf-8")

	resp, err = c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	results, err := io.ReadAll(resp.Body)

	PrintHeaders(resp.Header)
	c.PrintCookies(urlStr)

	fmt.Println("")
	fmt.Println("")
	fmt.Println(string(results))
	return nil
}

func MapToHeaders(req *http.Request, m map[string]string) {

	for k, v := range m {

		req.Header.Set(k, v)
	}
	return
}
func getHiddenValues(r *http.Response) (map[string]string, error) {
	defer r.Body.Close()

	valMap := make(map[string]string)
	valMap["__VIEWSTATE"] = ""
	valMap["__VIEWSTATEGENERATOR:"] = ""
	valMap["__EVENTVALIDATION:"] = ""
	valMap["__EVENTARGUMENT"] = ""
	valMap["__EVENTTARGET"] = ""

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	// find input tags with type=hidden attribute

	doc.Find("input").Each(func(index int, item *goquery.Selection) {
		inputType, _ := item.Attr("type")
		if inputType == "hidden" {
			name, _ := item.Attr("name")
			val, _ := item.Attr("value")
			valMap[name] = val
		}
	})

	return valMap, nil
}
func PrintHeaders(hdrs http.Header) {

	fmt.Println("****HEADERS****")
	for k, v := range hdrs {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("****END HEADERS****")

}
func PrintMap(m map[string]string) {
	fmt.Println("**** Map Values****")
	for k, v := range m {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("**** END Map Values****")

}
func (c *Client) PrintCookies(urlStr string) {

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Println("invalid URL: ", urlStr)
	}
	fmt.Println("****COOKIES****")

	for i, v := range c.Jar.Cookies(u) {
		log.Printf("%d: %s\n", i, v)
	}
	fmt.Println("****END COOKIES****")

}
