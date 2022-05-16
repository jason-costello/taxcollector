package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	_ "github.com/lib/pq"

	pdb "github.com/jason-costello/taxcollector/storage/pgdb"
)

// func Test_loadProxyList(t *testing.T) {
// 	type args struct {
// 		fileName string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    map[string]time.Time
// 		wantErr bool
// 	}{
// 		{
// 			name:    "valid file, valid return",
// 			args:    args{"proxies.txt"},
// 			want:    map[string]time.Time{},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := loadProxyList(tt.args.fileName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("loadProxyList() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("loadProxyList() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return true
}

var addrRegex = regexp.MustCompile(`(?P<num>\d+)\s(?P<street>.*)\s{2}(?P<city>.*),\s(?P<state>\w{2})\s(?P<zip>\d{5})`)

func Test_parseDetails(t *testing.T) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"127.0.0.1", 5432, "postgres", "password", "tax")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	pgdb := pdb.New(db)

	ctx := context.Background()

	var pr PropertyRecord
	property, err := pgdb.GetPropertyByID(ctx, 44712)
	if err != nil {
		t.Fatal(err)
	}
	pr.PropertyID = fmt.Sprint(property.ID)

	imp, err := pgdb.GetImprovementsByPropertyID(ctx, sql.NullInt32{Int32: property.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}
	var improvements []Improvement
	for _, x := range imp {
		im := Improvement{
			Name:        NullStringToString(x.Name),
			Description: NullStringToString(x.Description),
			StateCode:   NullStringToString(x.StateCode),
			LivingArea:  NullStringToString(x.LivingArea),
			Value:       fmt.Sprint(NullInt32ToInt(x.Value)),
			Details:     nil,
		}
		ids, err := pgdb.GetImprovementDetails(ctx, sql.NullInt32{Int32: x.ID, Valid: true})
		if err != nil {
			t.Fatal(err)
		}
		var improveDetails []ImprovDetail
		for _, sd := range ids {

			improveDetails = append(improveDetails, ImprovDetail{
				Type:         NullStringToString(sd.ImprovementType),
				Description:  NullStringToString(sd.ImprovementType),
				Class:        NullStringToString(sd.ImprovementType),
				ExteriorWall: NullStringToString(sd.ImprovementType),
				YearBuilt:    NullStringToString(sd.ImprovementType),
				SqFt:         NullStringToString(sd.ImprovementType),
			})

		}

		im.Details = improveDetails

		improvements = append(improvements, im)
	}

	pr.Improvements = improvements

	rollValues, err := pgdb.GetRollValuesByPropertyID(ctx, sql.NullInt32{Int32: property.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}
	var rvs []RollValue

	for _, r := range rollValues {

		rv := RollValue{
			Year:         fmt.Sprint(NullInt32ToInt(r.Year)),
			Improvements: fmt.Sprint(NullInt32ToInt(r.Improvements)),
			LandMarket:   fmt.Sprint(NullInt32ToInt(r.LandMarket)),
			AgValuation:  fmt.Sprint(NullInt32ToInt(r.AgValuation)),
			Appraised:    fmt.Sprint(NullInt32ToInt(r.Appraised)),
			HomesteadCap: fmt.Sprint(NullInt32ToInt(r.HomesteadCap)),
			Assessed:     fmt.Sprint(NullInt32ToInt(r.Assessed)),
		}

		rvs = append(rvs, rv)

	}
	pr.RollValue = rvs

	land, err := pgdb.GetLandByPropertyID(ctx, sql.NullInt32{Int32: property.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}

	lands := []Land{}
	for _, lnd := range land {
		lands = append(lands, Land{
			Number:      fmt.Sprint(NullInt32ToInt(lnd.Number)),
			Type:        NullStringToString(lnd.LandType),
			Description: NullStringToString(lnd.Description),
			Acres:       fmt.Sprint(NullFloat64ToFloat(lnd.Acres)),
			Sqft:        fmt.Sprint(NullFloat64ToFloat(lnd.SquareFeet)),
			EffFront:    fmt.Sprint(NullFloat64ToFloat(lnd.EffFront)),
			EffDepth:    fmt.Sprint(NullFloat64ToFloat(lnd.EffDepth)),
			MarketValue: fmt.Sprint(NullInt32ToInt(lnd.MarketValue)),
		})

	}

	pr.Land = lands

	juris, err := pgdb.GetJurisdictionsByPropertyID(ctx, sql.NullInt32{Int32: property.ID, Valid: true})
	if err != nil {
		t.Fatal(err)
	}

	var jurisdictions []TaxingJurisdiction

	for _, j := range juris {

		jurisdictions = append(jurisdictions, TaxingJurisdiction{
			Entity:         NullStringToString(j.Entity),
			Description:    NullStringToString(j.Description),
			TaxRate:        fmt.Sprint(NullInt32ToInt(j.TaxRate)),
			AppraisedValue: fmt.Sprint(NullInt32ToInt(j.AppraisedValue)),
			TaxableValue:   fmt.Sprint(NullInt32ToInt(j.TaxableValue)),
			EstimatedTax:   fmt.Sprint(NullInt32ToInt(j.EstimatedTax)),
		})

	}
	pr.Jurisdictions = jurisdictions

	b, err := json.Marshal(pr)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))

}
func NullFloat64ToFloat(s sql.NullFloat64) float64 {
	if s.Valid {
		return s.Float64
	}
	return float64(0.0)
}
func NullInt32ToInt(s sql.NullInt32) int {
	if s.Valid {
		return int(s.Int32)
	}
	return 0
}
func NullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// func Test_ProxyLoad(t *testing.T) {
// 	db, err := sql.Open("sqlite3", "./foo.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
//
// 	proxies, err := loadProxyList("proxies.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if err := CreateProxyTable(db); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if err := LoadProxyTable(db, proxies); err != nil {
// 		t.Fatal(err)
// 	}
//
// }
