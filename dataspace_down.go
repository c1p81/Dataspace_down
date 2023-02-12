// Author Luca Innocenti
// Florence 02/2023

// Dataspace Copernicus API
// https://documentation.dataspace.copernicus.eu/#/APIs/OData

// ************** ACCESS TOKEN
//curl -s -X POST "https://identity.dataspace.copernicus.eu/auth/realms/CDSE/protocol/openid-connect/token" \
//     -H "Content-Type: application/x-www-form-urlencoded" \
//     -d "username=<LOGIN>" \
//     -d 'password=<PASSWORD>' \
//     -d 'grant_type=password' \
//     -d 'client_id=cdse-public'|jq .access_token|tr -d '"'

// ************** QUERY IMAGES
//https://catalogue.dataspace.copernicus.eu/odata/v1/Products?$filter=Attributes/OData.CSC.DoubleAttribute/any(att:att/Name eq 'cloudCover' and att/OData.CSC.DoubleAttribute/Value le 40.00) and ContentDate/Start gt 2022-01-01T00:00:00.000Z and ContentDate/Start lt 2022-01-03T00:00:00.000Z&$top=10

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//  ################## JSON SEARCH STRUCTURE

type messaggio_dati struct {
	eodata int32 `json:"@odata.context"`
	ND     []struct {
		mediaContentType string `json:"@odata.mediaContentType"`
		Id               string `json:"Id"`
		Name             string `json:"Name"`
		ContentType      string `json:"ContentType"`
		ContentLength    string `json:"ContentLength"`
		OriginDate       string `json:"OriginDate"`
		ModificationDate string `json:"ModificationDate"`
		Online           bool   `json:"Online"`
		EvictionDate     string `json:"EvictionDate"`
		S3Path           string `json:"S3Path"`
		Footprint        string `json:"Footprint"`
		ContentDate      string `json:"ContentDate"`
		Checksum         string `json:"Checksum"`
	} `json:"value"`
}

func main() {
	var start_date string
	var end_date string
	var search_point_lat string
	var search_point_lon string
	var collection string
	var dest_path string
	var username string
	var password string
	var cloudCover string
	var ptype string
	var download bool

	tempo_e := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	tempo_s := time.Now().AddDate(0, 0, -5).UTC().Format("2006-01-02T15:04:05.000Z")

	flag.StringVar(&end_date, "end_date", tempo_e, "End sensing date (default today) Format YYYY-MM-DDThh:mm:ss.000Z")
	flag.StringVar(&start_date, "start_date", tempo_s, "Start sensing date (default today - 5 days) Format YYYY-MM-DDThh:mm:ss.000Z")
	flag.StringVar(&search_point_lon, "search_point_lon", "11.287615415088597", "Longitude")
	flag.StringVar(&search_point_lat, "search_point_lat", "43.78186592737776", "Latitude")
	flag.StringVar(&collection, "collection", "SENTINEL-2", "Collection")
	flag.StringVar(&dest_path, "dest_path", "./", "Download folder")
	flag.StringVar(&cloudCover, "cloudCover", "10", "Less than % cloud cover No Sentine-1")
	flag.StringVar(&username, "username", "", "Username (required)")
	flag.StringVar(&password, "password", "", "Password (required)")
	flag.StringVar(&ptype, "ptype", "", "Product type (default S2MSI2A)")
	flag.BoolVar(&download, "download", false, "If true start the download")

	flag.Parse()

	if (*&password == "") || (*&username == "") {
		fmt.Println("Set username and password")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	if (*&ptype == "") && (*&collection == "SENTINEL-2") {
		ptype = "S2MSI2A"
	}

	if (*&ptype == "") && (*&collection == "SENTINEL-1") {
		ptype = "GRD"
	}

	collection_list := map[string]bool{
		"SENTINEL-1":  true,
		"SENTINEL-2":  true,
		"SENTINEL-3":  true,
		"SENTINEL-5P": true,
	}

	S1_list := map[string]bool{
		"GRD": true,
		"SLC": true,
		"OCN": true,
		"RAW": true,
	}

	S2_list := map[string]bool{
		"S2MSI1C": true,
		"S2MSI2A": true,
	}

	if !collection_list[collection] {
		fmt.Println("Wrong collection name")
		fmt.Println("Avalaible SENTINEL-1, SENTINEL-2, SENTINEL-3, SENTINEL-5P")
		os.Exit(1)
	}

	if collection == "SENTINEL-1" {
		if !S1_list[ptype] {
			fmt.Println("Wrong product type for collection SENTINEL-1")
			fmt.Println("Avalaible SLC, GRD, RAW, OCN")
			os.Exit(1)
		}
	}

	if collection == "SENTINEL-2" {
		if !S2_list[ptype] {
			fmt.Println("Wrong product type for collection SENTINEL-2")
			fmt.Println("Avalaible S2MSI1C, S2MSI2A")
			os.Exit(1)
		}
	}

	fmt.Println("============== Parameters ======================")
	fmt.Println("Collection : " + collection)
	fmt.Println("Prod.Type  : " + ptype)
	fmt.Println("Start date : " + start_date)
	fmt.Println("End date   : " + end_date)
	fmt.Println("Latitude   : " + search_point_lat)
	fmt.Println("Longitude  : " + search_point_lon)
	fmt.Println("% Cloud    : less than " + cloudCover + " %")
	fmt.Println("Folder     : " + dest_path)
	fmt.Println("====================================")

	var url_search string
	if collection == "SENTINEL-1" {
		url_search = "https://catalogue.dataspace.copernicus.eu/odata/v1/Products?$filter=Attributes/OData.CSC.StringAttribute/any(att:att/Name%20eq%20%27productType%27%20and%20att/OData.CSC.StringAttribute/Value%20eq%20%27" + ptype + "%27)%20and%20Collection/Name%20eq%20%27" + collection + "%27%20and%20ContentDate/Start%20gt%20" + start_date + "%20and%20ContentDate/Start%20lt%20" + end_date + "%20and%20OData.CSC.Intersects(area=geography%27SRID=4326;POINT(" + search_point_lon + "%20" + search_point_lat + ")%27)"
	}
	if collection == "SENTINEL-2" {
		url_search = "https://catalogue.dataspace.copernicus.eu/odata/v1/Products?$filter=Attributes/OData.CSC.StringAttribute/any(att:att/Name%20eq%20%27productType%27%20and%20att/OData.CSC.StringAttribute/Value%20eq%20%27" + ptype + "%27)%20and%20Attributes/OData.CSC.DoubleAttribute/any(att:att/Name%20eq%20%27cloudCover%27%20and%20att/OData.CSC.DoubleAttribute/Value%20le%20" + cloudCover + ")%20and%20Collection/Name%20eq%20%27" + collection + "%27%20and%20ContentDate/Start%20gt%20" + start_date + "%20and%20ContentDate/Start%20lt%20" + end_date + "%20and%20OData.CSC.Intersects(area=geography%27SRID=4326;POINT(" + search_point_lon + "%20" + search_point_lat + ")%27)"
	}

	if collection == "SENTINEL-3" {
		url_search = "https://catalogue.dataspace.copernicus.eu/odata/v1/Products?$filter=Collection/Name%20eq%20%27" + collection + "%27%20and%20ContentDate/Start%20gt%20" + start_date + "%20and%20ContentDate/Start%20lt%20" + end_date + "%20and%20OData.CSC.Intersects(area=geography%27SRID=4326;POINT(" + search_point_lon + "%20" + search_point_lat + ")%27)"
	}

	if collection == "SENTINEL-5P" {
		url_search = "https://catalogue.dataspace.copernicus.eu/odata/v1/Products?$filter=Collection/Name%20eq%20%27" + collection + "%27%20and%20ContentDate/Start%20gt%20" + start_date + "%20and%20ContentDate/Start%20lt%20" + end_date + "%20and%20OData.CSC.Intersects(area=geography%27SRID=4326;POINT(" + search_point_lon + "%20" + search_point_lat + ")%27)"
	}

	//fmt.Println(url_search)
	resp, err := http.Get(url_search)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Search Error")
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
		fmt.Println("No response from remote server")
		return

	}

	//************* GET ACCESS TOKEN *****************************
	client := &http.Client{}

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("grant_type", "password")
	data.Set("client_id", "cdse-public")

	r, _ := http.NewRequest("POST", "https://identity.dataspace.copernicus.eu/auth/realms/CDSE/protocol/openid-connect/token", strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, err = client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	ro, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jdata map[string]interface{}
	err = json.Unmarshal([]byte(ro), &jdata)
	if err != nil {
		panic(err)
	}
	endpoint, ok := jdata["access_token"].(string)
	if !ok {
		panic("Error in access token")
	}

	// ************** START DOWNLOAD

	var mes messaggio_dati
	json.Unmarshal(body, &mes)
	for _, str := range mes.ND {
		if err != nil {
			panic("Error in JSON response")
		}
		if download {
			fmt.Println("Download " + str.Name)
			cmd_string := "curl -H \"Authorization: Bearer " + endpoint + "\" 'https://catalogue.dataspace.copernicus.eu/odata/v1/Products(" + str.Id + ")/$value' --output " + dest_path + str.Name + ".zip"
			//fmt.Println(cmd_string)
			cmd := exec.Command("/bin/sh", "-c", cmd_string)

			//_, err := cmd.Output()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()

			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("====================================")
			fmt.Println("====================================")

		} else {
			fmt.Println("Found : " + str.Name)
		}
	}
}
