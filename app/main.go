package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"goZBoard/configuration"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ZkbRequestTimeFormat = "200601021500"
	ZkbKillTimeFormat    = "2006-01-02 15:04:05"
)

var nilTime = (time.Time{}).UnixNano()

type KillData struct {
	KillID    int         `json:"KillID"`
	KillTime  EveKillTime `json:"KillTime"`
	SystemID  int         `json:"SystemID"`
	Victim    Victim      `json:"victim"`
	Attackers []Attacker  `json:"attackers"`
	Zkb       Zkb         `json:"zkb"`
	FinalBlow Attacker
}

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
	ShipTypeID      int     `json:"shipTypeID"`
}

type Victim struct {
	ShipTypeID      int    `json:"shipTypeID"`
	CharacterID     int    `json:"characterID"`
	CharacterName   string `json:"characterName"`
	CorporationID   int    `json:"corporationID"`
	CorporationName string `json:"corporationName"`
	FactionID       int    `json:"factionID"`
	FactionName     string `json:"factionName"`
	DamageTaken     int    `json:"damageTaken"`
}

type Zkb struct {
	TotalValue float64 `json:"totalValue"`
	Points     int     `json:"points"`
}

type EveKillTime struct {
	time.Time
}

func (ct *EveKillTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ZkbKillTimeFormat, s)
	return
}

func (ct *EveKillTime) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(ZkbKillTimeFormat))), nil
}

func (ct *EveKillTime) IsSet() bool {
	return ct.UnixNano() != nilTime
}

func main() {
	configFile := flag.String("config", "", "Configuration file")
	flag.Parse()
	configuration.LoadConfiguration(*configFile)

	var decoder *json.Decoder
	var killsMap []map[string]interface{}
	var reader io.ReadCloser

	startTime := time.Now().UTC().Add(-24 * time.Hour).Format(ZkbRequestTimeFormat)
	ticker := time.NewTicker(time.Second * 10)

	client := new(http.Client)
	url := fmt.Sprintf("%scorporationID/%s/startTime/%s/",
		configuration.Cfg.ZKillBoardUrl,
		configuration.Cfg.CorporationID,
		startTime)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("Request error: ", err)
		os.Exit(1)
	}

	request.Header.Add("Accept-Encoding", "gzip")
	request.Header.Add("User-Agent", "Gozboard Maintainer:Igor igor.intellij.kolinko@gmial.com")

	for {
		select {
		case <-ticker.C:
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Do request error: ", err)
				break
			}

			if response.StatusCode != 200 {
				fmt.Println("Invalid ressponse StatusCode: ", response.StatusCode)
				break
			}

			defer response.Body.Close()

			// Check that the server actually sent compressed data
			switch response.Header.Get("Content-Encoding") {

			case "gzip":
				reader, err = gzip.NewReader(response.Body)
				if err != nil {
					fmt.Println("Reader create err: ", err)
				}
			default:
				reader = response.Body
			}

			defer reader.Close()

			decoder = json.NewDecoder(reader)
			err = decoder.Decode(&killsMap)

			if err != nil {
				fmt.Println("Decode err: ", err)
				break

			}

			kills := []KillData{}
			for _, k := range killsMap {
				jsonkill, err := json.Marshal(k)
				if err != nil {
					fmt.Println("JSON Marshal error: ", err)
					break
				}
				var kill KillData
				err = json.Unmarshal(jsonkill, &kill)
				if err != nil {
					fmt.Println("JSON Unmarshal error: ", err)
					break
				}
				for _, lastHit := range kill.Attackers {
					if lastHit.FinalBlow == 1 {
						kill.FinalBlow = lastHit
						break
					}
				}
				kills = append(kills, kill)
			}
		}
	}

}
