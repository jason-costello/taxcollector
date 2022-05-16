package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Improvement struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	StateCode   string         `json:"stateCode,omitempty"`
	LivingArea  string         `json:"livingArea,omitempty"`
	Value       string         `json:"value,omitempty"`
	Details     []ImprovDetail `json:"details,omitempty"`
}

type ImprovDetail struct {
	Type         string `json:"type,omitempty"`
	Description  string `json:"description,omitempty"`
	Class        string `json:"class,omitempty"`
	ExteriorWall string `json:"exteriorWall,omitempty"`
	YearBuilt    string `json:"yearBuilt,omitempty"`
	SqFt         string `json:"sqFt,omitempty"`
}

func getImprovements(doc *goquery.Document) []Improvement {

	var improvements []Improvement
	var improvement Improvement
	doc.Find("#improvementBuildingDetails").Each(func(index int, div *goquery.Selection) {

		doc.Find("table").Each(func(tblIndex int, table *goquery.Selection) {
			var improvementDetails []ImprovDetail
			tblClass := table.AttrOr("class", "")
			if tblClass == "improvements" {
				improvement = getImprovement(table)
			}

			if tblClass == "improvementDetails" {
				improvementDetails = getImprovementDetail(table)

			}
			if improvementDetails != nil {

				improvement.Details = append(improvementDetails, improvementDetails...)
				// fmt.Printf("\nafter adding details for imprvment: %v  %#+v\n", improvement.Name, improvement.Details)
			}
			if improvement.Name != "" {

				improvements = append(improvements, improvement)
			}
		})
	})

	// fmt.Println(" ")
	// fmt.Printf("Returning: %#+v\n\n", improvements)
	return improvements
}

func getImprovement(tbl *goquery.Selection) Improvement {

	var improvement Improvement

	tbl.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
		row.Find("th,td").Each(func(cellIndex int, cell *goquery.Selection) {
			switch cellIndex {

			case 0:
				improvement.Name = strings.TrimSpace(cell.Text())
			case 1:
				improvement.Description = strings.TrimSpace(cell.Text())
			case 3:
				improvement.StateCode = strings.TrimSpace(cell.Text())
			case 5:
				improvement.LivingArea = strings.TrimSpace(cell.Text())
			case 7:
				improvement.Value = strings.TrimSpace(cell.Text())

			}
		})
	})

	return improvement

}

func getImprovementDetail(tbl *goquery.Selection) []ImprovDetail {
	improvementDetails := []ImprovDetail{}

	tbl.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {

		if rowIndex != 0 {
			var detail ImprovDetail
			row.Find("th,td").Each(func(cellIndex int, cell *goquery.Selection) {
				switch cellIndex {
				case 1:
					detail.Type = strings.TrimSpace(cell.Text())
				case 2:
					detail.Description = strings.TrimSpace(cell.Text())
				case 3:
					detail.Class = strings.TrimSpace(cell.Text())
				case 4:
					detail.ExteriorWall = strings.TrimSpace(cell.Text())
				case 5:
					detail.YearBuilt = strings.TrimSpace(cell.Text())
				case 6:
					detail.SqFt = strings.TrimSpace(cell.Text())

				}

			})
			if detail.Description != "" {
				improvementDetails = append(improvementDetails, detail)
			}
		}
	})
	return improvementDetails
}
