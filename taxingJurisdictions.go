package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TaxingJurisdiction struct {
	Entity         string `json:"entity,omitempty"`
	Description    string `json:"description,omitempty"`
	TaxRate        string `json:"taxRate,omitempty"`
	AppraisedValue string `json:"appraisedValue,omitempty"`
	TaxableValue   string `json:"taxableValue,omitempty"`
	EstimatedTax   string `json:"estimatedTax,omitempty"`
}

func getTaxingJurisdictions(doc *goquery.Document) []TaxingJurisdiction {

	var taxingJurisdictions []TaxingJurisdiction
	doc.Find("#taxingJurisdictionDetails > table.tableData").Each(func(index int, table *goquery.Selection) {
		table.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
			var taxJur TaxingJurisdiction
			row.Find("td").Each(func(cellIndex int, cell *goquery.Selection) {
				switch cellIndex {

				case 0:
					taxJur.Entity = strings.TrimSpace(strings.TrimSpace(cell.Text()))
				case 1:
					taxJur.Description = strings.TrimSpace(strings.TrimSpace(cell.Text()))
				case 2:
					taxJur.TaxRate = strings.TrimSpace(strings.TrimSpace(cell.Text()))
				case 3:
					taxJur.AppraisedValue = strings.TrimSpace(strings.TrimSpace(cell.Text()))
				case 4:
					taxJur.TaxableValue = strings.TrimSpace(strings.TrimSpace(cell.Text()))
				case 5:
					taxJur.EstimatedTax = strings.TrimSpace(strings.TrimSpace(cell.Text()))

				default:
				}

			})
			if taxJur.Entity != "" {
				taxingJurisdictions = append(taxingJurisdictions, taxJur)
			}
		})

	})
	return taxingJurisdictions

}
