package main

type sStatsInfo struct {
          timestamp string `json:"timestamp"`
          value float64 `json:"value"`
          limit float64 `json:"limit"`
          filename string  `json:"filename"`
        }
type sContainerAlert struct {
        Type  string `json:"type"`
        Timestamp string `json:"timestamp"`
        ContainerUuid string `json:"container_uuid"`
        EnvironmentId string `json:"environment_id"`
        ContainerName string `json:"container_name"`
        Namespace string `json:"namespace"`
        MaxValue float64  `json:"max_value"`
        MinValue float64  `json:"min_value"`
        AvgValue float64  `json:"avg_value"`
        TargetValue float64 `json:"target_value"`
        TriggeredAlerts bool `json:"triggerd_alerts"`
        AlertStartTime string 
        AlertEndTime string
        AlertMessage string `json:"alert_message"`
        Stats []sStatsInfo  `json:"stats"`
        
     
    }