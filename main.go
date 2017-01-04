package main

/*
Generate slack token: https://api.slack.com/web

Environment Variables:
  * SLACK_API_TOKEN
  * SLACK_ROOM
  * HIPCHAT_API_TOKEN
  * HIPCHAT_ROOM_ID
  * HIPCHAT_SERVER (optional)

*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/bluele/slack"
	flag "github.com/ogier/pflag"
	"github.com/tbruyelle/hipchat-go/hipchat"
	yaml "gopkg.in/yaml.v2"
	"net/http"
)

type Trigger struct {
	Operator string
	Value    int64
}

type Notifier struct {
	Name string
}

type TriggeredAlert struct {
	Hash        string
	TriggeredAt time.Time
}

type Alert struct {
	Name         string
	Type         string
	Hash         string
	Function     string
	Limit        int
	Timeshift    string
	GroupBy      string `yaml:"group_by"`
	Query        string
	Interval     float64
	Trigger      Trigger
	NotifiersRaw []string   `yaml:"notifiers"`
	Notifiers    []Notifier `yaml:"-"`
}

var httpClient http.Client

var (
	tMutex                sync.Mutex
	triggeredAlerts       = map[string]TriggeredAlert{}
)



func main() {
	var file *string = flag.StringP("config", "c", "", "Config file to use")
	flag.Parse()

	setupInflux()

	alerts := []Alert{}

	data, _ := ioutil.ReadFile(*file)
	err := yaml.Unmarshal(data, &alerts)
	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%+v\n", alerts)
	}

	setupHttpClient()

	done := make(chan bool)
	for _, alert := range alerts {
		go func(alert Alert) {
			alert.Setup()
			for {
				alert.Run()
				time.Sleep(time.Duration(alert.Interval) * time.Second)
			}
		}(alert)
	}
	<-done // wait
}
