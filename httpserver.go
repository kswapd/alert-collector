package main

import (
	"strconv"
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
	//"math"
)
/*
const container_cpu = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`		//ok
const container_memory = `select value from "container_memory_usage_bytes"`										//ok	
const container_network_tx = `select derivative(value,1s) from "container_network_transmit_bytes_total"`		//ok
const container_network_rx = `select derivative(value,1s) from "container_network_receive_bytes_total"`			//ok*/
//const container_disk = `select *  from /.*container_filesystem_capacity_\/*|.*container_filesystem_usage_\/*/`		//todo
//const mysql_connection = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`		
//const redis_hits = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
//const redis_memory = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`
//const nginx_accept = `select difference(value)/elapsed(value) from "container_cpu_usage_seconds_total"`


var (
	listenPort, _ = strconv.Atoi(getEnv("LISTEN_PORT", "8011"))
	ruleUrl = getEnv("RULE_URL", "http://127.0.0.1:8077/alert/v1/rule")
)

type MetricsJson struct {
	Key        string `json:"key"`
	Condition  string `json:"condition"`
	Value      float64 `json:"value"`
	QueryRaw   string `json:"queryRaw"`
	RuleStatus string `json:"ruleStatus"`
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
		if conMetrics.RuleStatus == "0" {
			continue
		}
		var _alert Alert
		_alert.Name = "container_" + conMetrics.Key
		_alert.Function = "average"
		_alert.GroupBy = "container_uuid"
		_alert.Interval = 60
		_alert.Limit = 500
		_alert.Timeshift = "10m"
		_alert.Query = conMetrics.QueryRaw
		if conMetrics.Key == "network_tx" || conMetrics.Key == "network_rx" {
			_alert.Trigger.Value = conMetrics.Value * 1000000
		} else {
			_alert.Trigger.Value  = conMetrics.Value

		}
		_alert.Type = "influxdb"
		_alert.Trigger.Operator = conMetrics.Condition

		_alert.NotifiersRaw = []string{"sendAlert"}
		alertRule = append(alertRule, _alert)
	}
	for _, appJson := range ruleJson.App {
		for _, appMetrics := range appJson.AppParam {
			if appMetrics.RuleStatus == "0" {
				continue
			}
			var _appAlert Alert
			_appAlert.Name = appJson.AppType + "_" + appMetrics.Key
			_appAlert.Function = "average"
			_appAlert.GroupBy = "container_uuid"
			_appAlert.Interval = 60
			_appAlert.Limit = 500
			_appAlert.Timeshift = "10m"
			_appAlert.Type = "influxdb"
			_appAlert.Trigger.Operator = appMetrics.Condition
			_appAlert.Trigger.Value = appMetrics.Value
			_appAlert.NotifiersRaw = []string{"sendAlert"}
			alertRule = append(alertRule, _appAlert)
		}
	}
}




