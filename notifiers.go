package main

/*
Generate slack token: https://api.slack.com/web
*/

import (
	"fmt"
	"net/url"
	"os"
	pagerduty "github.com/PagerDuty/go-pagerduty"
	"github.com/bluele/slack"
	"github.com/fatih/color"
	"github.com/tbruyelle/hipchat-go/hipchat"
	"net/http"
	"net"
	"time"
	"encoding/json"
	"log"
	"bytes"
	"io/ioutil"
)

type ParamJson struct {
	PKey   string `json:"p_key"`
	PValue string `json:"p_value"`
}

type AlertInfoJson struct {
	Status        string `json:"status"`
	AlertType     string `json:"alert_type"` //M-监控 L-日志
	AlertDim      string `json:"alert_dim"`  //C-容器 A-应用
	AppType       string `json:"app_type"`
	Msg           string `json:"msg"`
	EnvironmentId string `json:"environment_id"`
	ContainerUuid string `json:"container_uuid"`
	ContainerName string `json:"container_name"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Namespace     string `json:"namespace"`
	Data          []ParamJson `json:"data"`
}

const method = "POST"

var alertUrl = ""

type AlertData struct {
	AlertData []AlertInfoJson `json:"alert_data"`
}

func (this *Notifier) sendAlert(alertData AlertData) {
	sendBody, errParse := json.Marshal(alertData)
	if errParse != nil {
		log.Fatalln("Parse the alertdata error..")
		return
	}
	req, err := http.NewRequest(method, alertUrl, bytes.NewReader(sendBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("the data is %v", string(data))
	defer resp.Body.Close()
}

func setupHttpClient() {
	alertUrl = getEnv("ALERT_URL", "http://54.223.149.108:8077/alert/v1/info/receive")
	httpClient = http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second * 20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

