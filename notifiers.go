package main

/*
Generate slack token: https://api.slack.com/web
*/

import (
	"fmt"
	"os"
	"net/http"
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

const METHOD_POST = "POST"

var alertUrl = "http://54.222.160.114:8082/alert/v1/info/receive"

type AlertData struct {
	AlertInfo []AlertInfoJson `json:"alert_data"`
}

func (this *Notifier) sendAlert(alertData AlertData) {
	fmt.Printf("the alert data is %+v\n", alertData)
	sendBody, errParse := json.Marshal(alertData)
	if errParse != nil {
		log.Fatalln("Parse the alertdata error..")
		return
	}
	fmt.Printf("the string is %s\n", string(sendBody))
	client := &http.Client{}
	req, err := http.NewRequest(METHOD_POST, alertUrl, bytes.NewReader(sendBody))
	if err != nil {
		log.Fatalln("http post err", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("http post err", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("read response body error", err)
	}
	fmt.Printf("the data is %s.\n", string(data))
	defer resp.Body.Close()
}

func (this *Notifier) Run(msg string, isNotifier bool) {
	fmt.Println("the alert msg run is %s", msg)
}

func setupHttpClient() {
	alertUrl = getEnv("ALERT_URL", "http://54.223.149.108:8077/alert/v1/info/receive")
	/*httpClient = http.Client{
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
	}*/
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
