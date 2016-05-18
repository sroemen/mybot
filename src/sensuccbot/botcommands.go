package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	bcmds["silence"] = cmds{"silence", []string{"server", "client/alert", "duration=[15m|30m|8h|24h|forever]"}}
}

func isValidCommand(cmd string) bool {
	for _, c := range bcmds {
		if c.Name == strings.ToLower(cmd) {
			log.Printf("found %s", cmd)
			return true
		}
	}
	return false
}

func isValidHashCommand(m []string) bool {
	c := m[0]
	if strings.HasPrefix(c, "#") {
		//log.Printf("testing command %s ", c[1:])
		return isValidCommand(c[1:])
		//} else {
		//	log.Println("no # command")
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
	} else if cmd == "silence" {
		if snoozealert(args) {
			return "Alert silenced"
		} else {
			return "Alert not silenced"
		}
	}

	return ""
}

func listalerts(args []string) []string {
	rs := make([]string, 0)
	v := []events1{}
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
					if status != "undef" {
						rs = append(rs, fmt.Sprintf("server: %s check: %s Severity: %s Silenced?: %v", dta.Server, check, status, (dta.CheckSilenced || dta.ClientSilenced)))
					}
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

type salerts struct {
	Server   string
	Client   string
	Duration int64
}

func snoozealert(args []string) bool {

	if len(args) < 2 {
		log.Printf("Too few arguments")
		return false
	}
	//log.Printf("%v args passed in. %#v", len(args), args)

	bodypost := salerts{}

	if len(args) > 2 {
		bodypost.Server = args[0]
		bodypost.Client = args[1]
		d, err := strconv.Atoi(args[2])
		if err != nil {
			log.Printf("Error converting string to int in snoozealert() %v", err)
			return false
		}
		bodypost.Duration = int64(d)
	} else {
		bodypost.Server = args[0]
		bodypost.Client = args[1]
		bodypost.Duration = 900
	}

	bdy, err := json.Marshal(bodypost)
	if err != nil {
		log.Printf("Error in snoozealert() Marshal() %v", err)
		return false
	}

	//log.Printf("posted body: %s (%#v)", bodypost, bdy)

	res, err := http.NewRequest("POST", "https://localhost:8091/api/latest/sensuSilence", bytes.NewBuffer(bdy))
	if err != nil {
		log.Printf("Error in snoozealert(): %v", err)
		return false
	}
	res.Header.Add("Content-Type", "application/json")

	response, err := client.Do(res)
	if err == nil {
		defer response.Body.Close()

		if response.StatusCode == 200 {
			/*
				data, err := ioutil.ReadAll(response.Body)
				if err == nil {
					json.Unmarshal(data, &v)
				} else {
					log.Printf("Failed to decode stash data (%v)", err)
					return rs
				}
			*/
			return true
		}
	}
	return false
}
