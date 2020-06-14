package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

type symbolInfo struct {
	ID   int    `json:"instrument_token"`
	Name string `json:"symbol"`
}

const (
	apiKey    string = "Your-API-Key"
	apiSecret string = "Your-API-Secret"
)

var symbols []symbolInfo
var kc *kiteconnect.Client //kite connect client
var requestToken string
var session kiteconnect.UserSession
var err error

func readConfig() {

	jsonfile, err := os.Open("nifty50.json")

	defer jsonfile.Close()

	if err != nil {
		println("failed to open file ", err)
	}

	byteVal, _ := ioutil.ReadAll(jsonfile)
	json.Unmarshal(byteVal, &symbols)

}

func login() bool {
	// Create a new Kite connect instance
	kc = kiteconnect.New(apiKey)

	// Login URL from which request token can be obtained
	fmt.Println("Open the following url in your browser:\n", kc.GetLoginURL())

	// Obtain request token after Kite Connect login flow
	// Run a temporary server to listen for callback
	srv := &http.Server{Addr: ":8888"}
	go http.HandleFunc("/login/", func(w http.ResponseWriter, r *http.Request) {
		requestToken = r.URL.Query()["request_token"][0]
		log.Println("request token", requestToken)
		go srv.Shutdown(context.TODO())
		w.Write([]byte("login successful!"))
		return
	})

	srv.ListenAndServe()

	// Get user details and access token
	session, err = kc.GenerateSession(requestToken, apiSecret)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return false
	}

	// Set access token
	kc.SetAccessToken(session.AccessToken)
	log.Println("session.AccessToken", session.AccessToken)
	return true
}

var from time.Time
var to time.Time

func fetchSymbolData(symbolID int, symbolName string, destFolder string) {

	//fetch from 01-Jan-2015 till last month
	//func Date(year int, month Month, day int, hour int, min int, sec int, nsec int, loc *Location) Time
	from = time.Date(2015, 1, 1, 9, 00, 0, 0, time.Now().Location())
	to = time.Date(2015, 1, 31, 15, 35, 0, 0, time.Now().Location())

	//Iterate for each month until 'from' is not equal to current month
	for !((from.Year() == time.Now().Year()) && (from.Month() == time.Now().Month())) {

		//Fetch Historical Date from Kite API
		HistoricalData, err := kc.GetHistoricalData(symbolID, "minute", from, to, false)

		if err != nil {
			log.Println("err:" + err.Error())
			return
		}

		//log.Println("Fetch Success. Now, saving to json file")

		if len(HistoricalData) != 0 {
			// build the file <symbol name><Month><year> convention
			fileName := destFolder + symbolName + from.Format(" 01-2006") + ".json"
			log.Println("fileName:" + fileName)

			jsonFile, _ := os.Create(fileName)
			defer jsonFile.Close()
			e, _ := json.Marshal(HistoricalData)
			jsonFile.Write(e)
		}

		from = from.AddDate(0, 1, 0)
		to = to.AddDate(0, 1, 0)

	}

}

func main() {

	//First login to Kite
	if login() == false {
		log.Println("login failed")
		return

	}

	//Create Root folder
	rootFolder := "./" + string(time.Now().Format("2006-Jan-02_15.04.05")) + "/"
	os.Mkdir(rootFolder, os.ModePerm)

	//read nifty50.json
	readConfig()

	for i := 0; i < len(symbols); i++ {
		fmt.Println("Fetching Name: ", symbols[i].Name)
		symbolFolder := rootFolder + symbols[i].Name + "/"

		os.Mkdir(symbolFolder, os.ModePerm)

		fetchSymbolData(symbols[i].ID, symbols[i].Name, symbolFolder)

	}

}
