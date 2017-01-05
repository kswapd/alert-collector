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
	"flag"
	yaml "gopkg.in/yaml.v2"
	"net/http"
)

type Trigger struct {
	Operator string
	Value    float64
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

var (
	tMutex sync.Mutex
	triggeredAlerts = map[string]TriggeredAlert{}
	httpClient http.Client
	alertRule []Alert
)

var (
	alertFile = flag.String("config_file", "example.yml", "Config alert file to use")
	influxAddr = flag.String("influx_addr", "54.223.73.138:8086", "host:port")
)

func main() {
	//var file *string = flag.StringP("config", "c", "", "Config file to use")

	flag.Parse()
	fmt.Printf("alert file:%s, influx address:%s\n", *alertFile, *influxAddr)
	setupInflux()

	alerts := []Alert{}

	data, _ := ioutil.ReadFile(*alertFile)
	err := yaml.Unmarshal(data, &alerts)
	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%+v\n", alerts)
	}

	setupHttpClient()
	getRules()

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
