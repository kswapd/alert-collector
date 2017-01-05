package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	//"strconv"
	"time"
	"github.com/influxdata/influxdb/client"
)

func queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: "containerdb",//os.Getenv("INFLUX_DB"),
	}
	if response, err := con.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	}
	return
}

func query(query string, alertId string) map[string] *sContainerAlert {
	ret := []float64{}

	//containerInfo := 
	
	
	res, err := queryDB(query)
	if err != nil {
		log.Fatal(err)
	}
	if len(res) < 1 {
		return nil
	}

	if len(res[0].Series) < 1 {
		return nil
	}

	//fmt.Printf("%#v.\n",res[0]);
	
	for index := 0; index < len(res[0].Series); index++ {
		se := res[0].Series[index]

//se.tags.container_uuid


		if _, ok := containerStatsInfo[alertId][se.Tags["container_uuid"]]; !ok {

			containerStatsInfo[alertId][se.Tags["container_uuid"]] = new(sContainerAlert)
			containerStatsInfo[alertId][se.Tags["container_uuid"]].TriggeredAlerts = false

		}
		containerStatsInfo[alertId][se.Tags["container_uuid"]].ContainerUuid = se.Tags["container_uuid"]
		containerStatsInfo[alertId][se.Tags["container_uuid"]].Type = "container-cpu"
		for i, row := range se.Values {
			t, err := time.Parse(time.RFC3339, row[0].(string))
			if err != nil {
				log.Fatal(err)
			}
			if row[1] == nil {
				continue
			}
			//fmt.Println(row)
			val, _ := row[1].(json.Number).Float64()
			ret = append(ret, val)


			/*if(se.Values[valIndex][2] == nil){			//todo , remove hard code.
				continue
			}*/
			timeStr := fmt.Sprintf("%s", row[0])
			valStr,err := row[1].(json.Number).Float64()
			var sinfo sStatsInfo
			sinfo.value = valStr
			sinfo.timestamp = timeStr

			containerStatsInfo[alertId][se.Tags["container_uuid"]].Stats = append(containerStatsInfo[alertId][se.Tags["container_uuid"]].Stats, sinfo)
			//containerStatsInfo[se.Name].Stats[i].value = valStr
			//containerStatsInfo[se.Name].Stats[i].timestamp = timeStr
			containerStatsInfo[alertId][se.Tags["container_uuid"]].Timestamp = timeStr
			if os.Getenv("DEBUG") == "true" {
				log.Printf("[%2d] %s: %d\n", i, t.Format(time.Stamp), val)
			}
		}

		//fmt.Printf("%#v.\n",containerStatsInfo[se.Tags["container_uuid"]]);
	}
	
	//fmt.Printf("%#v.\n",containerStatsInfo);
	return containerStatsInfo[alertId]
}

var con *client.Client

func setupInflux() {
	//influx_port, _ := strconv.ParseInt(os.Getenv("INFLUX_PORT"), 10, 0)

	u, err := url.Parse(fmt.Sprintf("http://%s", *influxAddr))
	if err != nil {
		log.Fatal(err)
	}

	conf := client.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PASS"),
	}

	con, err = client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	dur, ver, err := con.Ping()
	if err != nil {
		log.Fatal(err)
	}
	if os.Getenv("DEBUG") == "true" {
		log.Printf("Connected in %v | Version: %s", dur, ver)
	}
}
