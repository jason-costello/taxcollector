package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CoordsResponse struct {
	Info struct {
		Statuscode int `json:"statuscode"`
		Copyright  struct {
			Text         string `json:"text"`
			ImageUrl     string `json:"imageUrl"`
			ImageAltText string `json:"imageAltText"`
		} `json:"copyright"`
		Messages []interface{} `json:"messages"`
	} `json:"info"`
	Options struct {
		MaxResults        int  `json:"maxResults"`
		ThumbMaps         bool `json:"thumbMaps"`
		IgnoreLatLngInput bool `json:"ignoreLatLngInput"`
	} `json:"options"`
	Results []struct {
		ProvidedLocation struct {
			Location string `json:"location"`
		} `json:"providedLocation"`
		Locations []struct {
			Street             string `json:"street"`
			AdminArea6         string `json:"adminArea6"`
			AdminArea6Type     string `json:"adminArea6Type"`
			AdminArea5         string `json:"adminArea5"`
			AdminArea5Type     string `json:"adminArea5Type"`
			AdminArea4         string `json:"adminArea4"`
			AdminArea4Type     string `json:"adminArea4Type"`
			AdminArea3         string `json:"adminArea3"`
			AdminArea3Type     string `json:"adminArea3Type"`
			AdminArea1         string `json:"adminArea1"`
			AdminArea1Type     string `json:"adminArea1Type"`
			PostalCode         string `json:"postalCode"`
			GeocodeQualityCode string `json:"geocodeQualityCode"`
			GeocodeQuality     string `json:"geocodeQuality"`
			DragPoint          bool   `json:"dragPoint"`
			SideOfStreet       string `json:"sideOfStreet"`
			LinkId             string `json:"linkId"`
			UnknownInput       string `json:"unknownInput"`
			Type               string `json:"type"`
			LatLng             struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"latLng"`
			DisplayLatLng struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"displayLatLng"`
			MapUrl string `json:"mapUrl"`
		} `json:"locations"`
	} `json:"results"`
}

func main() {
	db, err := sql.Open("sqlite3", "../foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "select id, address from properties where (latitude is null OR longitude is null) and address != '' ;"

	propertyMap := make(map[string]string)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	for rows.Next() {

		var addr, id string
		if err := rows.Scan(&id, &addr); err != nil {
			panic(err)
		}
		fmt.Printf("id=%s   addr=%s\n", id, addr)

		if addr == "" {
			continue
		}

		propertyMap[id] = addr
	}

	for propertyID, address := range propertyMap {
		time.Sleep(800 * time.Millisecond)
		c := http.DefaultClient

		url := fmt.Sprintf(`http://open.mapquestapi.com/geocoding/v1/address?key=cOzLbu1ePZ5tWvTYrJjad4yXmasrr1mk&location=%s`, address)

		fmt.Println("url: ", url)
		resp, err := c.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var cr CoordsResponse
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(b, &cr); err != nil {
			continue
		}

		var lat, lon float64
		for _, c := range cr.Results {
			for _, l := range c.Locations {
				fmt.Println(l.GeocodeQuality)
				fmt.Println(l.GeocodeQualityCode)
				fmt.Println(l.LatLng.Lat)
				fmt.Println(l.LatLng.Lng)
				fmt.Println(address)
				lat = l.LatLng.Lat
				lon = l.LatLng.Lng
			}
		}

		query := "update properties set latitude = $1, longitude = $2 where id = $3"
		if _, err := db.Exec(query, lat, lon, propertyID); err != nil {
			fmt.Println(err)

		}
	}
}
