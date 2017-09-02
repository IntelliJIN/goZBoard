package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	cfg "goZBoard/configuration"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	_ "net/http/pprof"
	"log"
)

const (
	zkbRequestTimeFormat = "200601021500"
	zkbKillTimeFormat    = "2006-01-02 15:04:05"
	gZIP                 = "gzip"
)

var conf = cfg.NewConfig()

var nilTime = (time.Time{}).UnixNano()

//KillData representation of kill from zKillboard
type KillData struct {
	KillID    int         `json:"KillID"`
	KillTime  EveKillTime `json:"KillTime"`
	Victim    Victim      `json:"victim"`
	Attackers []Attacker  `json:"attackers"`
	Zkb       Zkb         `json:"zkb"`
	FinalBlow Attacker
	System
}

//System representation of EVE System
type System struct {
	SystemID   int `json:"solarSystemID"`
	SystemName string
}

//Attacker representation of zKillboard Attacker character
type Attacker struct {
	CharacterID     int     `json:"characterID"`
	CharacterName   string  `json:"characterName"`
	CorporationID   int     `json:"corporationID"`
	CorporationName string  `json:"corporationName"`
	AllianceID      int     `json:"allianceID"`
	AllianceName    string  `json:"allianceName"`
	FactionID       int     `json:"factionID"`
	FactionName     string  `json:"factionName"`
	SecurityStatus  float64 `json:"securityStatus"`
	DamageDone      int     `json:"damageDone"`
	FinalBlow       int     `json:"finalBlow"`
	WeaponTypeID    int     `json:"weaponTypeID"`
	Ship
}

//Victim representation of zKillboard victim character
type Victim struct {
	Ship
	CharacterID     int    `json:"characterID"`
	CharacterName   string `json:"characterName"`
	CorporationID   int    `json:"corporationID"`
	CorporationName string `json:"corporationName"`
	FactionID       int    `json:"factionID"`
	FactionName     string `json:"factionName"`
	DamageTaken     int    `json:"damageTaken"`
}

//Zkb representation of zKillboard loose value
type Zkb struct {
	TotalValue float64 `json:"totalValue"`
	Points     int     `json:"points"`
}

//Ship representation of EVE Ship
type Ship struct {
	ShipTypeID int `json:"shipTypeID"`
	ShipName   string
}

//getShipName get ship name from CCP site
func (s *Ship) getShipName() *Ship {
	//TODO make some kind of Ship Cache
	var decoder *json.Decoder
	var shipInfo map[string]interface{}
	var reader io.ReadCloser
	URL := fmt.Sprintf("https://esi.tech.ccp.is/latest/universe/types/%[1]v/?datasource=tranquility&language=en-us",
		s.ShipTypeID)
	requestHeader := map[string]string{
		"Accept-Encoding": gZIP,
		"Accept":          "application/json",
		"User-Agent":      "goZBoard Maintainer:Igor igor.intellij.kolinko@gmial.com",
	}
	response, err := sendGetRequest(requestHeader, URL)
	if err != nil {
		fmt.Println("Do request error: ", err)
	}

	if response.StatusCode != 200 {
		fmt.Println("Invalid ressponse StatusCode: ", response.StatusCode)
	}

	defer func() {
		err = response.Body.Close()
		checkError(err)
	}()
	// Check that the server actually sent compressed data
	switch response.Header.Get("Content-Encoding") {

	case gZIP:
		reader, err = gzip.NewReader(response.Body)
		checkError(err)
	default:
		reader = response.Body
	}

	defer func() {
		err = reader.Close()
		checkError(err)
	}()
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&shipInfo)
	checkError(err)
	name, ok := shipInfo["name"].(string)
	if ok {
		s.ShipName = name
	} else {
		s.ShipName = ""
	}
	return s
}

//sendGetRequest function get a map of request.Header params and URL to make a request
func sendGetRequest(args map[string]string, url string) (*http.Response, error) {
	client := new(http.Client)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range args {
		request.Header.Add(key, value)
	}
	response, err := client.Do(request)
	return response, err
}

func (s *System) getSystemName() *System {
	//TODO make some kind of System Cache
	var decoder *json.Decoder
	var shipInfo map[string]interface{}
	var reader io.ReadCloser
	URL := fmt.Sprintf("https://esi.tech.ccp.is/latest/universe/systems/%[1]v/?datasource=tranquility&language=en-us",
		s.SystemID)
	requestHeader := map[string]string{
		"Accept-Encoding": gZIP,
		"Accept":          "application/json",
		"User-Agent":      "goZBoard Maintainer:Igor igor.intellij.kolinko@gmial.com",
	}
	response, err := sendGetRequest(requestHeader, URL)
	if err != nil {
		fmt.Println("Do request error: ", err)
	}

	if response.StatusCode != 200 {
		fmt.Println("Invalid ressponse StatusCode: ", response.StatusCode)
	}

	defer func() {
		err = response.Body.Close()
		checkError(err)
	}()
	// Check that the server actually sent compressed data
	switch response.Header.Get("Content-Encoding") {

	case gZIP:
		reader, err = gzip.NewReader(response.Body)
		checkError(err)
	default:
		reader = response.Body
	}

	defer func() {
		err = reader.Close()
		checkError(err)
	}()
	decoder = json.NewDecoder(reader)
	err = decoder.Decode(&shipInfo)
	checkError(err)
	name, ok := shipInfo["name"].(string)
	if ok {
		s.SystemName = name
	} else {
		s.SystemName = ""
	}
	return s
}

//EveKillTime representation of time received from zKillboard
type EveKillTime struct {
	time.Time
}

//UnmarshalJSON representation of json.Unmarshal interface for non-standard zKillboard kill time
func (ct *EveKillTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(zkbKillTimeFormat, s)
	return
}

//MarshalJSON representation of json.Unmarshal interface for non-standard zKillboard kill time
func (ct *EveKillTime) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(zkbKillTimeFormat))), nil
}

//IsSet representation of json interface for non-standard zKillboard kill time
func (ct *EveKillTime) IsSet() bool {
	return ct.UnixNano() != nilTime
}

func main() {
	configFile := flag.String("config", "", "Configuration file")
	flag.Parse()
	conf.LoadConfiguration(*configFile)

	killChan := make(chan KillData)
	var decoder *json.Decoder
	var killsMap []map[string]interface{}
	var reader io.ReadCloser

	go sendToSlack(killChan)
	startTime := time.Now().UTC().Add(-72 * time.Hour).Format(zkbRequestTimeFormat)
	ticker := time.NewTicker(time.Second * 10)

	url := fmt.Sprintf("%scorporationID/%s/startTime/%s/",
		conf.ZKillBoardAPIURL,
		conf.CorporationID,
		startTime)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	for {
		select {
		case <-ticker.C:
			requestHeader := map[string]string{
				"Accept-Encoding": gZIP,
				"User-Agent":      "goZBoard Maintainer:Igor igor.intellij.kolinko@gmial.com",
			}
			response, err := sendGetRequest(requestHeader, url)
			if err != nil {
				fmt.Println("Do request error: ", err)
				break
			}

			if response.StatusCode != 200 {
				fmt.Println("Invalid ressponse StatusCode: ", response.StatusCode)
				break
			}

			defer func() {
				err = response.Body.Close()
				checkError(err)
			}()

			// Check that the server actually sent compressed data
			switch response.Header.Get("Content-Encoding") {
			case gZIP:
				reader, err = gzip.NewReader(response.Body)
				checkError(err)
			default:
				reader = response.Body
			}

			defer func() {
				err = reader.Close()
				checkError(err)
			}()
			decoder = json.NewDecoder(reader)
			err = decoder.Decode(&killsMap)
			checkError(err)

			for _, k := range killsMap {
				jsonkill, err := json.Marshal(k)
				checkError(err)
				var kill KillData
				err = json.Unmarshal(jsonkill, &kill)
				checkError(err)
				for _, lastHit := range kill.Attackers {
					if lastHit.FinalBlow == 1 {
						kill.FinalBlow = lastHit
						break
					}
				}
				killChan <- kill
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("ERROR: ", err)
		panic(err)
	}
}

//sendToSlack send processed kills info to Slack channel from config
func sendToSlack(killChan chan KillData) {
	for kill := range killChan {
		slackMSG := map[string]interface{}{}
		killData := map[string]interface{}{}

		url := fmt.Sprintf("%[1]vkill/%[2]v/", conf.ZKillBoardURL, kill.KillID)
		killData["title_link"] = url
		killData["thumb_url"] = fmt.Sprintf("https://imageserver.eveonline.com/Render/%[1]v_64.png",
			kill.Victim.ShipTypeID)
		killData["fallback"] = url
		killData["text"] = ""
		killData["color"] = "danger"
		killData["pretext"] = fmt.Sprintf("Death <%[1]vcharacter/%[2]v|%[3]v>",
			conf.ZKillBoardURL,
			kill.Victim.CharacterID,
			kill.Victim.CharacterName)

		killInfo := map[string]interface{}{
			"title": kill.KillTime,
			"value": "Loose",
			"short": false,
		}

		victimInfo := map[string]interface{}{
			"title": "Pilot:",
			"value": fmt.Sprintf("Name: <%[1]vcharacter/%[2]v|%[3]v> "+
				"Corporation: <%[1]vcorporation/%[4]v|%[5]v>",
				conf.ZKillBoardURL,
				kill.Victim.CharacterID,
				kill.Victim.CharacterName,
				kill.Victim.CorporationID,
				kill.Victim.CorporationName),
			"short": false,
		}

		shipInfo := map[string]interface{}{
			"title": "Ship:",
			"value": fmt.Sprintf("Name: <%[1]vship/%[2]v|%[3]v> "+
				"System: <%[1]vsystem/%[4]v|%[5]v> "+
				"Damage taken: %[6]v "+
				"Points: %[7]v",
				conf.ZKillBoardURL,
				kill.Victim.ShipTypeID,
				kill.Victim.getShipName().ShipName,
				kill.SystemID,
				kill.getSystemName().SystemName,
				kill.Victim.DamageTaken,
				kill.Zkb.Points),
			"short": false,
		}

		totalAttackers := map[string]interface{}{
			"title": "Attackers: ",
			"value": fmt.Sprintf("Pilots involved [%[1]v] | "+
				"Last Shot by: <%[2]vcharacter/%[3]v|%[4]v> | "+
				"Corporation: <%[2]vcorporation/%[5]v|%[6]v>",
				len(kill.Attackers),
				conf.ZKillBoardURL,
				kill.FinalBlow.CharacterID,
				kill.FinalBlow.CharacterName,
				kill.FinalBlow.CorporationID,
				kill.FinalBlow.CorporationName),
			"short": false,
		}

		value := map[string]interface{}{
			"title": "Value Lost",
			"value": fmt.Sprintf("%.2f ISK", kill.Zkb.TotalValue),
			"short": false,
		}

		data := []map[string]interface{}{}
		data = append(data, killInfo, shipInfo, victimInfo, totalAttackers, value)
		killData["fields"] = data
		dd := []map[string]interface{}{}
		dd = append(dd, killData)
		slackMSG["attachments"] = dd
		marshaledData, err := json.Marshal(slackMSG)
		checkError(err)
		resp, err := http.Post(conf.WebHookURL,
			"application/x-www-form-urlencoded",
			bytes.NewBuffer(marshaledData))
		checkError(err)

		defer func() {
			err = resp.Body.Close()
			checkError(err)
		}()
		b, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			fmt.Printf("Error sending Slackbot command %d: %s",
				resp.StatusCode, string(b))
		}
	}
}
