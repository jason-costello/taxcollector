package tax

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jason-costello/taxcollector/storage/pgdb"
)

type Land struct {
	Number      string `json:"number,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Acres       string `json:"acres,omitempty"`
	Sqft        string `json:"sqft,omitempty"`
	EffFront    string `json:"effFront,omitempty"`
	EffDepth    string `json:"effDepth,omitempty"`
	MarketValue string `json:"marketValue,omitempty"`
}

func getLandInfo(doc *goquery.Document) []Land {

	var lands []Land
	doc.Find("#landDetails > table").Each(func(index int, table *goquery.Selection) {
		var land Land
		table.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
			row.Find("td").Each(func(cellIndex int, cell *goquery.Selection) {
				switch cellIndex {

				case 0:
					land.Number = strings.TrimSpace(cell.Text())
				case 1:
					land.Type = strings.TrimSpace(cell.Text())
				case 2:
					land.Description = strings.TrimSpace(cell.Text())
				case 3:
					land.Acres = strings.TrimSpace(cell.Text())
				case 4:
					land.Sqft = strings.TrimSpace(cell.Text())
				case 5:
					land.EffFront = strings.TrimSpace(cell.Text())
				case 6:
					land.EffDepth = strings.TrimSpace(cell.Text())
				case 7:
					land.MarketValue = strings.TrimSpace(cell.Text())
				default:
				}
			})
			if land.Number != "" {
				lands = append(lands, land)
			}

		})
	})

	return lands
}

func NullFloat64ToString(f sql.NullFloat64) string {
	if f.Valid {
		return fmt.Sprint(f.Float64)
	}
	return ""
}

func FromLandDBModel(land []pgdb.Land) []Land {

	var ll []Land
	for _, l := range land {
		ll = append(ll, Land{
			Number:      NullInt32ToString(l.Number),
			Type:        NullStringToString(l.LandType),
			Description: NullStringToString(l.Description),
			Acres:       NullFloat64ToString(l.Acres),
			Sqft:        NullFloat64ToString(l.SquareFeet),
			EffFront:    NullFloat64ToString(l.EffFront),
			EffDepth:    NullFloat64ToString(l.EffDepth),
			MarketValue: NullInt32ToString(l.MarketValue),
		})
	}
	return ll
}
