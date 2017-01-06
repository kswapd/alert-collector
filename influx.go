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
	"strings"
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



func indexOf(strs []string, dst string) int {
	for k, v := range strs {
		if v == dst {
			return k
		}
	}
	return -1 //未找到dst，返回-1
}
func queryTags(query string, containerStatsInfo map[string] *sContainerAlert) map[string] *sContainerAlert {
	
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



		//timeInd:= indexOf(ret[0].Series[0].Columns, "time")
		uuidInd := indexOf(res[0].Series[0].Columns, "container_uuid")
		envIdInd := indexOf(res[0].Series[0].Columns, "environment_id")
		nameInd := indexOf(res[0].Series[0].Columns, "container_name")
		namespaceInd := indexOf(res[0].Series[0].Columns, "namespace")
		typeInd := indexOf(res[0].Series[0].Columns, "type")

		//containerMonitorTag.Timestamp = fmt.Sprintf("%s", ret[0].Series[0].Values[0][timeInd])
		Container_uuid := fmt.Sprintf("%s", res[0].Series[0].Values[0][uuidInd])
		Environment_id := fmt.Sprintf("%s", res[0].Series[0].Values[0][envIdInd])
		Container_name := fmt.Sprintf("%s", res[0].Series[0].Values[0][nameInd])
		Namespace := fmt.Sprintf("%s", res[0].Series[0].Values[0][namespaceInd])
		Type := fmt.Sprintf("%s", res[0].Series[0].Values[0][typeInd])


		containerStatsInfo[Container_uuid].EnvironmentId = Environment_id
		containerStatsInfo[Container_uuid].ContainerName = Container_name
		containerStatsInfo[Container_uuid].Namespace = Namespace
		containerStatsInfo[Container_uuid].Type = Type

	return containerStatsInfo
}


func queryDiskUsage(query string, containerStatsInfo map[string] *sContainerAlert)  map[string] *sContainerAlert {
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

		valInd := indexOf(res[0].Series[0].Columns, "value")
		timeInd := indexOf(res[0].Series[0].Columns, "time")
		if _, ok := containerStatsInfo[se.Tags["container_uuid"]]; !ok {

			containerStatsInfo[se.Tags["container_uuid"]] = new(sContainerAlert)
			containerStatsInfo[se.Tags["container_uuid"]].TriggeredAlerts = false

		}
		containerStatsInfo[se.Tags["container_uuid"]].ContainerUuid = se.Tags["container_uuid"]
		containerStatsInfo[se.Tags["container_uuid"]].Type = "container-cpu"
		for i, row := range se.Values {
			t, err := time.Parse(time.RFC3339, row[timeInd].(string))
			if err != nil {
				log.Fatal(err)
			}
			if row[valInd] == nil {
				continue
			}
			//fmt.Println(row)
			val, _ := row[valInd].(json.Number).Float64()
			ret = append(ret, val)


			/*if(se.Values[valIndex][2] == nil){			//todo , remove hard code.
				continue
			}*/
			timeStr := fmt.Sprintf("%s", row[timeInd])
			valStr,err := row[valInd].(json.Number).Float64()

			fsIndex := -1
			var sinfo sStatsInfo			
			sinfo.timestamp = timeStr
			if strings.Contains(se.Name, "container_filesystem_capacity_") {
				sinfo.filename = strings.TrimPrefix(se.Name, "container_filesystem_capacity_")
				sinfo.limit = valStr

				for i, v := range containerStatsInfo[se.Tags["container_uuid"]].Stats {
					if v.filename == sinfo.filename && v.timestamp == sinfo.timestamp{
					   fsIndex = i
	              	   containerStatsInfo[se.Tags["container_uuid"]].Stats[i].limit = sinfo.limit
					   break
					}
				}

				if fsIndex == -1 {
					containerStatsInfo[se.Tags["container_uuid"]].Stats = append(containerStatsInfo[se.Tags["container_uuid"]].Stats, sinfo)
				}

				continue

			}else if strings.Contains(se.Name, "container_filesystem_usage_"){
				sinfo.filename = strings.TrimPrefix(se.Name, "container_filesystem_usage_")
				sinfo.value = valStr

				for i, v := range containerStatsInfo[se.Tags["container_uuid"]].Stats {
					if v.filename == sinfo.filename && v.timestamp == sinfo.timestamp{
					   fsIndex = i
	              	   containerStatsInfo[se.Tags["container_uuid"]].Stats[i].value = sinfo.value
					   break
					}
				}

				if fsIndex == -1 {
					containerStatsInfo[se.Tags["container_uuid"]].Stats = append(containerStatsInfo[se.Tags["container_uuid"]].Stats, sinfo)
				}

				continue

			}

			
			//containerStatsInfo[se.Tags["container_uuid"]].Stats = append(containerStatsInfo[se.Tags["container_uuid"]].Stats, sinfo)




			//containerStatsInfo[se.Name].Stats[i].value = valStr
			//containerStatsInfo[se.Name].Stats[i].timestamp = timeStr
			containerStatsInfo[se.Tags["container_uuid"]].Timestamp = timeStr
			if os.Getenv("DEBUG") == "true" {
				log.Printf("[%2d] %s: %d\n", i, t.Format(time.Stamp), val)
			}
		}

		//fmt.Printf("%#v.\n",containerStatsInfo[se.Tags["container_uuid"]]);
	}
	
	//fmt.Printf("%#v.\n",containerStatsInfo);
	return containerStatsInfo
}




func query(query string, containerStatsInfo map[string] *sContainerAlert)  map[string] *sContainerAlert {
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


		if _, ok := containerStatsInfo[se.Tags["container_uuid"]]; !ok {

			containerStatsInfo[se.Tags["container_uuid"]] = new(sContainerAlert)
			containerStatsInfo[se.Tags["container_uuid"]].TriggeredAlerts = false
			containerStatsInfo[se.Tags["container_uuid"]].AlertMessage = "alert message"

		}
		containerStatsInfo[se.Tags["container_uuid"]].ContainerUuid = se.Tags["container_uuid"]
		containerStatsInfo[se.Tags["container_uuid"]].Type = "container-cpu"
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

			containerStatsInfo[se.Tags["container_uuid"]].Stats = append(containerStatsInfo[se.Tags["container_uuid"]].Stats, sinfo)
			//containerStatsInfo[se.Name].Stats[i].value = valStr
			//containerStatsInfo[se.Name].Stats[i].timestamp = timeStr
			containerStatsInfo[se.Tags["container_uuid"]].Timestamp = timeStr
			if os.Getenv("DEBUG") == "true" {
				log.Printf("[%2d] %s: %d\n", i, t.Format(time.Stamp), val)
			}
		}

		//fmt.Printf("%#v.\n",containerStatsInfo[se.Tags["container_uuid"]]);
	}
	
	//fmt.Printf("%#v.\n",containerStatsInfo);
	return containerStatsInfo
}

var con *client.Client

func setupInflux() {
	//influx_port, _ := strconv.ParseInt(os.Getenv("INFLUX_PORT"), 10, 0)

	u, err := url.Parse(fmt.Sprintf("http://%s", influxAddr))
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
