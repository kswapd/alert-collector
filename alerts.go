package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"time"
	"strings"
	"github.com/fatih/color"
	//"log"
)

func (alert *Alert) ApplyFunction(orginValue map[string]*sContainerAlert) map[string]*sContainerAlert {
	var appliedFunction float64

	for _, info := range orginValue {

		if strings.Contains(alert.Name, "disk") {
			for i, _ := range info.Stats {
					if info.Stats[i].limit > 0{
						info.Stats[i].value = info.Stats[i].value/info.Stats[i].limit
					}else{
						info.Stats[i].value = 0.0
					}
				}
		}
		if len(info.Stats) > 0 {
				appliedFunction = info.Stats[0].value
		}

		if alert.Function == "average" {
			for _, s := range info.Stats {
				appliedFunction += float64(s.value)
			}

			if(len(info.Stats) < 1){

				fmt.Println("err alert length")
				fmt.Printf("%v.\n",info);
				appliedFunction = 0.0

			}else{
				appliedFunction = appliedFunction / float64(len(info.Stats))
			}
			info.AvgValue = appliedFunction
			info.TargetValue = appliedFunction
		} else if alert.Function == "max" {
			for _, s := range info.Stats {
				appliedFunction = math.Max(appliedFunction, s.value)
				info.MaxValue = appliedFunction
				info.TargetValue = appliedFunction
			}
		} else if alert.Function == "min" {
			for _, s := range info.Stats {
				appliedFunction = math.Min(appliedFunction, s.value)
				info.MinValue = appliedFunction
				info.TargetValue = appliedFunction
			}
		}
	}
	return orginValue
}

func (alert *Alert) Setup() {
	hash := md5.Sum([]byte(alert.Name))
	alert.Hash = hex.EncodeToString(hash[:])
	for _, n := range alert.NotifiersRaw {
		alert.Notifiers = append(alert.Notifiers, Notifier{Name: n})
	}
	alert.containerStatsInfo = make(map[string] *sContainerAlert)

}

func (alert *Alert) Run() {
	const RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	if os.Getenv("DEBUG") == "true" {
		fmt.Println("Query: ", fmt.Sprintf("%s limit %d", alert.Query, alert.Limit))
	}


	//fmt.Println(alert.Name)

	var infos map[string] *sContainerAlert
	groupByQuery := ""
	if len(alert.GroupBy) > 0 {
		groupByQuery = fmt.Sprintf("GROUP BY %s", alert.GroupBy)
	}
	finalQuery := fmt.Sprintf("%s where time > now() - %s %s limit %d",
		alert.Query, alert.Timeshift, groupByQuery, alert.Limit)

	fmt.Println(finalQuery)
	if strings.Contains(alert.Name, "disk") {
		infos = queryDiskUsage(finalQuery, alert.containerStatsInfo)
	} else {
		infos = query(finalQuery, alert.containerStatsInfo)
	}

	infos = alert.ApplyFunction(infos)

	if os.Getenv("DEBUG") == "true" {
		fmt.Println("Applied Func: ", infos)
	}

	var allAlert AlertData

	for uuid, info := range infos {

		alert_triggered := false
		switch alert.Trigger.Operator {
		case "GT":
			alert_triggered = info.TargetValue > float64(alert.Trigger.Value)
		case "LT":
			alert_triggered = info.TargetValue < float64(alert.Trigger.Value)
		case "GTE":
			alert_triggered = info.TargetValue >= float64(alert.Trigger.Value)
		case "LTE":
			alert_triggered = info.TargetValue <= float64(alert.Trigger.Value)
		}

		if alert_triggered {
			message := fmt.Sprintf("*[!] %s--%s triggered!* Value: %.2f | Trigger: %s %.2f %s",
				alert.Name, uuid, info.TargetValue, alert.Trigger.Operator, alert.Trigger.Value, info.AlertStartTime)
			color.Red(message)
			alertAlreadyTriggered := false
			tMutex.Lock()
			if info.TriggeredAlerts {
				color.Yellow(fmt.Sprintf("[already triggered at %s] %s", info.AlertStartTime, uuid))
				alertAlreadyTriggered = true
			} else {
				info.TriggeredAlerts = true
				info.AlertStartTime = info.Timestamp
			}
			tMutex.Unlock()
			if !alertAlreadyTriggered {
				/*for _, n := range alert.Notifiers {
					n.Run(message, true)
				}*/
				tagQuery := fmt.Sprintf("select * from container_cpu_usage_seconds_total where container_uuid='%s' order by time desc limit 1", uuid)

				 queryTags(tagQuery, alert.containerStatsInfo)


				var param ParamJson
				param.PKey = "cpu"
				param.PValue = fmt.Sprintf("%.2f", info.TargetValue)

				var alert AlertInfoJson
				alert.Status = "alert"
				alert.AlertType = "M"
				alert.AlertDim = "C"
				alert.AppType = "container"
				alert.Msg = info.AlertMessage
				alert.EnvironmentId = info.EnvironmentId
				alert.ContainerUuid = uuid
				alert.ContainerName = info.ContainerName
				alert.StartTime = info.AlertStartTime
				alert.EndTime = alert.StartTime
				st, err := time.Parse(RFC3339Nano, alert.StartTime)
				if err == nil {
					//queryValidation = false
					
					st = st.Add(time.Second*30)
					etStr := st.Format(RFC3339Nano)
					alert.EndTime = etStr


					st = st.Add(time.Second*(-60))
					stStr := st.Format(RFC3339Nano)
					alert.StartTime = stStr

					//fmt.Println("stStr:"+stStr+"es:"+etStr)


				}else{
					fmt.Println("Error: "+info.AlertStartTime)
				}
				

				
				alert.Namespace = info.Namespace
				alert.Data = append(alert.Data, param)
				allAlert.AlertInfo = append(allAlert.AlertInfo, alert)

				//sendAlert()
			}

		} else {
			tMutex.Lock()
			if info.TriggeredAlerts {
				info.TriggeredAlerts = false
				//message := fmt.Sprintf("*[+] %s--%s resolved * Value: %.2f | Trigger: %s %.2f",
				//alert.Name, uuid, info.TargetValue, alert.Trigger.Operator, alert.Trigger.Value)
				/*for _, n := range alert.Notifiers {
					n.Run(message, false)
				}*/
				color.Green("[+] %s - Alert resolved.", alert.Name)
			}
			tMutex.Unlock()
			//color.Green(fmt.Sprintf("[+] %s--%s passed. (%.2f)", alert.Name,uuid, info.TargetValue))
		}

	}

	if len(allAlert.AlertInfo) > 0 {
		for _, n := range alert.Notifiers {
			n.sendAlert(allAlert)
		}
	}

}
