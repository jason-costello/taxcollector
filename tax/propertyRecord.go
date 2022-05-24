package tax

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jason-costello/taxcollector/storage/pgdb"
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

func GetPropertyRecord(doc *goquery.Document) (PropertyRecord, error) {

	propertyRecord := PropertyRecord{}
	itemMap := loadPropertyDetailItems()
	for k, v := range itemMap {
		val := doc.Find(v.SelectorText).Text()
		v.Value = strings.TrimSpace(val)
		itemMap[k] = v
	}

	propertyRecord.MapscoMapID = itemMap["mapscoMapID"].Value
	propertyRecord.OwnershipPercentage = strings.Replace(itemMap["ownershipPercentage"].Value, "%", "", 1)
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

func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
func NullInt32ToInt32(ni sql.NullInt32) int32 {
	if ni.Valid {
		return ni.Int32
	}
	return 0
}
func NullInt32ToString(ni sql.NullInt32) string {
	if ni.Valid {
		return fmt.Sprint(ni.Int32)
	}
	return ""
}
func Int32ToString(i int32) string {
	return fmt.Sprint(i)
}
func FromPropertyDBModel(property pgdb.Property) PropertyRecord {

	return PropertyRecord{
		PropertyID:          Int32ToString(property.ID),
		OwnerID:             NullInt32ToString(property.OwnerID),
		OwnerName:           NullStringToString(property.OwnerName),
		OwnerMailingAddress: NullStringToString(property.OwnerMailingAddress),
		Zoning:              NullStringToString(property.Zoning),
		NeighborhoodCD:      NullStringToString(property.NeighborhoodCd),
		Neighborhood:        NullStringToString(property.Neighborhood),
		Address:             NullStringToString(property.Address),
		LegalDescription:    NullStringToString(property.LegalDescription),
		GeographicID:        NullStringToString(property.GeographicID),
		Exemptions:          NullStringToString(property.Exemptions),
		OwnershipPercentage: NullFloat64ToString(property.OwnershipPercentage),
		MapscoMapID:         NullStringToString(property.MapscoMapID),
		RollValue:           nil,
		Land:                nil,
		Improvements:        nil,
		Jurisdictions:       nil,
	}

}
