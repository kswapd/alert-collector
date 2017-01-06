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
	//"io/ioutil"
	"os"
	"sync"
	"time"
	"flag"
	//yaml "gopkg.in/yaml.v2"
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
	containerStatsInfo map[string] *sContainerAlert
}

var (
	tMutex sync.Mutex
	triggeredAlerts = map[string]TriggeredAlert{}
	httpClient http.Client
	alertRule []Alert
)

var (
	//alertFile = flag.String("config_file", "example.yml", "Config alert file to use")
	//alertFile = getEnv("config_file", "http://54.223.149.108:8077/alert/v1/info/receive")
	influxAddr = getEnv("INFLUX_ADDR", "54.223.73.138:8086")
	//containerStatsInfo = make(map[string]map[string] *sContainerAlert)
)

func main() {
	//var file *string = flag.StringP("config", "c", "", "Config file to use")

	flag.Parse()
	fmt.Printf("get influx address:%s\n", influxAddr)
	setupInflux()

	alerts := []Alert{}

	//data, _ := ioutil.ReadFile(*alertFile)
	//err := yaml.Unmarshal(data, &alerts)
	/*if err != nil {
		panic(err)
	}*/

	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("%+v\n", alerts)
	}

	setupHttpClient()
	getRules()

	done := make(chan bool)
	for _, alert := range alertRule {
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
