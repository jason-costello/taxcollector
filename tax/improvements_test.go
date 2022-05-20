package tax

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func Test_getImprovements(t *testing.T) {

	d, err := ioutil.ReadFile("test_data/2163.html")
	if err != nil {
		t.Fatal(err)
	}
	r := strings.NewReader(string(d))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatal(err)
	}
	improvements := getImprovements(doc)

	t.Logf("%#+v\n", improvements)

	if improvements == nil {
		t.Error()
	}
}
func Test_GetDetails(t *testing.T) {
	d, err := ioutil.ReadFile("test_data/2163.html")
	if err != nil {
		t.Fatal(err)
	}
	r := strings.NewReader(string(d))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatal(err)
	}
	doc.Find("#improvementBuildingDetails").Each(func(index int, div *goquery.Selection) {
		doc.Find("table").Each(func(tblIndex int, table *goquery.Selection) {
			tblClass := table.AttrOr("class", "")

			if tblClass == "improvements" {
				detailsTable := doc.Next()

				t.Log(detailsTable.Text())
				fmt.Printf("%#+v\n", getImprovementDetail(detailsTable))
			}
			if tblClass == "improvementDetails" {

			}
		})
	})

}
