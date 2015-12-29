package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	bcmds map[string]cmds
)

type cmds struct {
	Name string
	Args []string
}

func init() {
	bcmds = make(map[string]cmds, 0)

	bcmds["help"] = cmds{"help", []string{}}
	bcmds["listalerts"] = cmds{"listalerts", []string{"showall"}}
	bcmds["silence"] = cmds{"silence", []string{"alert", "duration=[15m|30m|8h|24h|forever]"}}
	//bcmds["resolve"] = cmds{"resolve", []string{"alert"}}

}

func isValidCommand(cmd string) bool {
	for _, c := range bcmds {
		if c.Name == strings.ToLower(cmd) {
			return true
		}
	}
	return false
}

func runCommand(cmd string, args []string) string {
	if cmd == "help" {
		r := "available commands: "
		for _, c := range bcmds {
			a := ""
			if len(c.Args) > 0 {
				a = strings.Join(c.Args, " ")
			}
			if len(a) > 0 {
				r = r + "\n" + c.Name + " arguments: " + a
			} else {
				r = r + "\n" + c.Name
			}
		}
		return r
	} else if cmd == "listalerts" {
		return strings.Join(listalerts(args), "\n")
	}

	return ""
}

func listalerts(args []string) []string {
	rs := make([]string, 0)
	v := []events{}
	sall := false
	if len(args) > 0 {
		if args[0] == "showall" {
			sall = true
		}
	}

	res, err := http.NewRequest("GET", "https://localhost:8091/api/latest/sensuAlarms", nil)
	if err != nil {
		log.Printf("Error in listalerts(): %v", err)
		return rs
	}

	response, err := client.Do(res)
	if err == nil {
		defer response.Body.Close()

		if response.StatusCode == 200 {
			data, err := ioutil.ReadAll(response.Body)
			if err == nil {
				json.Unmarshal(data, &v)
			} else {
				log.Printf("Failed to decode stash data (%v)", err)
				return rs
			}

			// do something with the data
			for _, dta := range v {
				if sall || !(dta.CheckSilenced || dta.ClientSilenced) {
					check := dta.Client.Name + "/" + dta.Check.Name
					status := "undef"

					switch dta.Check.Status {
					case 0:
						status = "Normal"
					case 1:
						status = "Warning"
					case 2:
						status = "Critical"
					case 3:
						status = "Informational"
					}
					rs = append(rs, fmt.Sprintf("check: %s Severity: %s Silenced?: %v", check, status, (dta.CheckSilenced || dta.ClientSilenced)))
				}
			}
			if len(rs) > 0 {
				return rs
			}
		} else {
			log.Printf("ERROR: listalerts() respoonse code %v", response.StatusCode)
			return rs
		}

	} else {
		log.Printf("ERROR: listalerts() error: %v", err)
		return rs
	}

	return []string{"No events found (try 'listalerts showall')"}
}
