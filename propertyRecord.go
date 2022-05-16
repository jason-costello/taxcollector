package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PropertyRecord struct {
	PropertyID          string               `json:"propertyID"`
	OwnerID             string               `json:"ownerID"`
	OwnerName           string               `json:"ownerName"`
	OwnerMailingAddress string               `json:"ownerMailingAddress"`
	Zoning              string               `json:"zoning"`
	NeighborhoodCD      string               `json:"neighborhoodCD"`
	Neighborhood        string               `json:"neighborhood"`
	Address             string               `json:"address"`
	LegalDescription    string               `json:"legalDescription"`
	GeographicID        string               `json:"geographicID"`
	Exemptions          string               `json:"exemptions"`
	OwnershipPercentage string               `json:"ownershipPercentage"`
	MapscoMapID         string               `json:"mapscoMapID"`
	RollValue           []RollValue          `json:"rollValue"`
	Land                []Land               `json:"land"`
	Improvements        []Improvement        `json:"improvements"`
	Jurisdictions       []TaxingJurisdiction `json:"jurisdictions"`
}

type PropertyDetailItem struct {
	Name         string `json:"name,omitempty"`
	Value        string `json:"value,omitempty"`
	SelectorText string `json:"selectorText,omitempty"`
}

func getPropertyRecord(doc *goquery.Document) (PropertyRecord, error) {

	propertyRecord := PropertyRecord{}
	itemMap := loadPropertyDetailItems()
	for k, v := range itemMap {
		val := doc.Find(v.SelectorText).Text()
		v.Value = strings.TrimSpace(val)
		itemMap[k] = v
	}

	propertyRecord.MapscoMapID = itemMap["mapscoMapID"].Value
	propertyRecord.OwnershipPercentage = itemMap["ownershipPercentage"].Value
	propertyRecord.Exemptions = itemMap["exemptions"].Value
	propertyRecord.GeographicID = itemMap["geographicID"].Value
	propertyRecord.LegalDescription = itemMap["legalDescription"].Value
	propertyRecord.Address = itemMap["address"].Value
	propertyRecord.Neighborhood = itemMap["neighborhood"].Value
	propertyRecord.NeighborhoodCD = itemMap["neighborhoodCD"].Value
	propertyRecord.PropertyID = itemMap["property"].Value
	propertyRecord.OwnerID = itemMap["ownerID"].Value
	propertyRecord.PropertyID = itemMap["propertyID"].Value
	propertyRecord.OwnerName = itemMap["ownerName"].Value
	propertyRecord.OwnerMailingAddress = itemMap["ownerMailingAddress"].Value
	propertyRecord.Zoning = itemMap["zoning"].Value

	propertyRecord.Improvements = getImprovements(doc)
	propertyRecord.Land = getLandInfo(doc)
	propertyRecord.Jurisdictions = getTaxingJurisdictions(doc)
	propertyRecord.RollValue = getRollValue(doc)

	return propertyRecord, nil
}

func loadPropertyDetailItems() map[string]PropertyDetailItem {

	detailItemMap := make(map[string]PropertyDetailItem)
	detailItemMap["propertyID"] = PropertyDetailItem{
		Name:         "propertyID",
		Value:        "",
		SelectorText: "#propertyDetails > table > tbody > tr:nth-child(2) > td:nth-child(2)",
	}

	detailItemMap["geographicID"] = PropertyDetailItem{
		Name:         "geographicID",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(3) > td:nth-child(2)`,
	}
	detailItemMap["legalDescription"] = PropertyDetailItem{
		Name:         "legalDescription",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(2) > td.propertyDetailsLegalDescription`,
	}
	detailItemMap["zoning"] = PropertyDetailItem{
		Name:         "zoning",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(3) > td:nth-child(4)`,
	}

	detailItemMap["address"] = PropertyDetailItem{
		Name:         "address",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(8) > td:nth-child(2)`,
	}
	detailItemMap["neighborhood"] = PropertyDetailItem{
		Name:         "neighborhood",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(9) > td:nth-child(2)`,
	}
	detailItemMap["neighborhoodCD"] = PropertyDetailItem{
		Name:         "neighborhoodCD",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(10) > td:nth-child(2)`,
	}
	detailItemMap["mapscoMapID"] = PropertyDetailItem{
		Name:         "mapscoMapID",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(9) > td:nth-child(4)`,
	}
	detailItemMap["ownerName"] = PropertyDetailItem{
		Name:         "ownerName",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(12) > td:nth-child(2)`,
	}
	detailItemMap["ownerMailingAddress"] = PropertyDetailItem{
		Name:         "ownerMailingAddress",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(13) > td:nth-child(2)`,
	}
	detailItemMap["ownerID"] = PropertyDetailItem{
		Name:         "ownerID",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(12) > td:nth-child(4)`,
	}
	detailItemMap["ownershipPercentage"] = PropertyDetailItem{
		Name:         "ownershipPercentage",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(13) > td:nth-child(4)`,
	}
	detailItemMap["exemptions"] = PropertyDetailItem{
		Name:         "exemptions",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(14) > td:nth-child(4)`,
	}
	detailItemMap["ownerMailingAddress"] = PropertyDetailItem{
		Name:         "ownerMailingAddress",
		Value:        "",
		SelectorText: `#propertyDetails > table > tbody > tr:nth-child(13) > td:nth-child(2)`,
	}

	return detailItemMap
}
