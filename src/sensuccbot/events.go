package main

// events struct contains data from the sensu event api
type events struct {
	Id          string `json:"id"`
	Client      Client `json:"client"`
	Check       Check  `json:"check"`
	Occurrences int    `json:"occurrences"`
	Action      string `json:"action"`
	//Sensu Control Center additions
	StatusText     string
	ShowLabel      bool
	StatusLabel    string
	EventId        string
	CheckSilenced  bool
	ClientSilenced bool
}

// Client struct contains data from the sensu client api
type Client struct {
	Name          string            `json:"name"`
	Address       string            `json:"address"`
	Subscriptions []string          `json:"subscriptions"`
	SafeMode      bool              `json:"safe_mode"`
	Keepalive     keepaliveSettings `json:"keepalive"`
	Version       string            `json:"version"`
	Timestamp     int64             `json:"timestamp"`
}

type Check struct {
	Handlers    []string `json:"handlers"`
	Command     string   `json:"command"`
	Interval    int      `json:"interval"`
	Subscribers []string `json:"subscribers"`
	Occurrences int      `json:"occurrences"`
	Refresh     int      `json:"refresh"`
	Name        string   `json:"name"`
	Issued      int64    `json:"issued"`
	Executed    int64    `json:"executed"`
	Duration    float32  `json:"duration"`
	Output      string   `json:"output"`
	Status      int      `json:"status"`
	History     []int    `json:"history"`
}

// struct for Client.Keepalive
type keepaliveSettings struct {
	Thresholds thresholdTypes `json:"thresholds"`
	Handlers   []string       `json:"handlers"`
	Refresh    int            `json:"refresh"`
}

// struct for Client.Keepalive.Thresholds
type thresholdTypes struct {
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
}
