package main

import (
	"strconv"
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

const container_cpu = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const container_memory = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const container_network_tx = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const container_network_rx = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const container_disk = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const mysql_connection = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const redis_hits = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const redis_memory = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
const nginx_accept = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`

var (
	listenPort, _ = strconv.Atoi(getEnv("LISTEN_PORT", "8011"))
	ruleUrl = getEnv("RULE_URL", "http://cloudos.hnaresearch.com/alert/v1/rule")
)

type MetricsJson struct {
	Key       string `json:"key"`
	Condition string `json:"condition"`
	Value     float64 `json:"value"`
}

type RuleJson struct {
	Container []MetricsJson `json:"container"`
	App       []AppJson `json:"app"`
}

type AppJson struct {
	AppType  string `json:"app_type"`
	AppParam []MetricsJson `json:"app_param"`
}

const METHOD_GET = "get"

type handle struct {

}

func (this *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("server is handle something...")
}

func StartServer() {
	h := &handle{}
	log.Println("listen:", listenPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), h)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

func getRules() {
	resp, err := http.Get(ruleUrl)
	if err != nil {
		log.Fatalln("get rules from alert error ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("read response body error...")
	}
	var ruleJson RuleJson
	errParse := json.Unmarshal(body, &ruleJson)
	if errParse != nil {
		log.Fatalln("body pares error ", errParse)
	}
	alertRule = alertRule[:0]
	for _, conMetrics := range ruleJson.Container {
		var _alert Alert
		_alert.Name = "container_" + conMetrics.Key
		_alert.Function = "average"
		_alert.GroupBy = "container_uuid"
		_alert.Interval = 60
		_alert.Limit = 500
		_alert.Timeshift = "10m"
		switch conMetrics.Key {
		case "cpu":
			_alert.Query = container_cpu
		case "memory":
			_alert.Query = container_memory
		case "disk":
			_alert.Query = container_disk
		case "network_tx":
			_alert.Query = container_network_tx
		case "network_rx":
			_alert.Query = container_network_rx
		default:
			log.Fatalln("no container metrics match....")
		}
		_alert.Type = "influxdb"
		_alert.Trigger.Operator = conMetrics.Condition
		_alert.Trigger.Value = conMetrics.Value
		alertRule = append(alertRule, _alert)
	}
	for _, appJson := range ruleJson.App {
		var _appAlert Alert
		for _, appMetrics := range appJson.AppParam {
			_appAlert.Name = appJson.AppType + "_" + appMetrics.Key
			_appAlert.Function = "average"
			_appAlert.GroupBy = "container_uuid"
			_appAlert.Interval = 60
			_appAlert.Limit = 500
			_appAlert.Timeshift = "10m"
			switch expr := appJson.AppType + "_" + appMetrics.Key; expr{
			case "mysql_connection":
				_appAlert.Query = mysql_connection
			case "redis_hits":
				_appAlert.Query = redis_hits
			case "redis_memory":
				_appAlert.Query = redis_memory
			case "nginx_accept":
				_appAlert.Query = nginx_accept
			default:
				log.Fatalln("no app metrics match...")
			}
			_appAlert.Type = "influxdb"
			_appAlert.Trigger.Operator = appMetrics.Condition
			_appAlert.Trigger.Value = appMetrics.Value
			alertRule = append(alertRule, _appAlert)
		}
	}
	fmt.Println("the rule is %+v\n", alertRule)
}



