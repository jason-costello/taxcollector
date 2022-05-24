package tax

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jason-costello/taxcollector/storage/pgdb"
)

type RollValue struct {
	Year         string `json:"year,omitempty"`
	Improvements string `json:"improvements,omitempty"`
	LandMarket   string `json:"landMarket,omitempty"`
	AgValuation  string `json:"agValuation,omitempty"`
	Appraised    string `json:"appraised,omitempty"`
	HomesteadCap string `json:"homesteadCap,omitempty"`
	Assessed     string `json:"assessed,omitempty"`
}

func getRollValue(doc *goquery.Document) []RollValue {

	// Roll Value History
	var rollValues []RollValue
	doc.Find("#rollHistoryDetails > table").Each(func(index int, table *goquery.Selection) {
		table.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
			var rollValue RollValue
			row.Find("td").Each(func(cellIndex int, cell *goquery.Selection) {
				switch cellIndex {

				case 0:
					rollValue.Year = strings.TrimSpace(cell.Text())
				case 1:
					rollValue.Improvements = strings.TrimSpace(cell.Text())
				case 2:
					rollValue.LandMarket = strings.TrimSpace(cell.Text())
				case 3:
					rollValue.AgValuation = strings.TrimSpace(cell.Text())
				case 4:
					rollValue.Appraised = strings.TrimSpace(cell.Text())
				case 5:
					rollValue.HomesteadCap = strings.TrimSpace(cell.Text())
				case 6:
					rollValue.Assessed = strings.TrimSpace(cell.Text())
				default:
				}
			})
			if rollValue.Year != "" {
				rollValues = append(rollValues, rollValue)
			}
		})

	})

	return rollValues
}
func FromRollValueDBModel(rollValue []pgdb.RollValue) []RollValue {

	var rv []RollValue

	for _, r := range rollValue {

		rv = append(rv, RollValue{
			Year:         NullInt32ToString(r.Year),
			Improvements: NullInt32ToString(r.Improvements),
			LandMarket:   NullInt32ToString(r.LandMarket),
			AgValuation:  NullInt32ToString(r.AgValuation),
			Appraised:    NullInt32ToString(r.Appraised),
			HomesteadCap: NullInt32ToString(r.HomesteadCap),
			Assessed:     NullInt32ToString(r.Assessed),
		})
	}
	return rv
}
