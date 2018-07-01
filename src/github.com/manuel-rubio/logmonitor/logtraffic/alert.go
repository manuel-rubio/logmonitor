package logtraffic

import(
    "fmt"
    "time"
    "encoding/json"

    "github.com/manuel-rubio/logmonitor/logyzer"
    "github.com/manuel-rubio/logmonitor/logtimer"
)

type TrafficAlert struct {
    threshold int
    thresholdCrossed bool
    hitsPerSecond []int
    hitsTotal int
    currentSecond int
}

func TrafficLoop(logs <-chan logyzer.LogEntry, done <-chan bool, threshold int) {
    timer := logtimer.New(1)
    traffic := TrafficAlert{
        threshold: threshold,
        thresholdCrossed: false,
        hitsPerSecond: make([]int, 60),
        currentSecond: time.Now().Second(),
    }
    for {
        select {
        case entry := <-logs:
            traffic = ProcessLog(entry, traffic)
        case <-timer.Tick():
            traffic = CheckTime(traffic)
        case <-done:
            return
        }
    }
}

func ProcessLog(entry logyzer.LogEntry, traffic TrafficAlert) (TrafficAlert) {
    // TODO: QUESTION: should I count the bad lines for traffic alert?
    if entry.IsBadLine() {
        return traffic
    }
    second := traffic.currentSecond
    traffic.hitsPerSecond[second] ++
    return Check(traffic)
}

func CheckTime(traffic TrafficAlert) (TrafficAlert) {
    second := time.Now().Second()
    if traffic.currentSecond != second {
        traffic.currentSecond = second
        traffic.hitsPerSecond[second] = 0
    }
    if traffic.thresholdCrossed {
        return Check(traffic)
    }
    return traffic
}

func Check(traffic TrafficAlert) (TrafficAlert) {
    traffic.hitsTotal = 0
    for _, value := range traffic.hitsPerSecond {
        traffic.hitsTotal += value
    }
    if traffic.hitsTotal >= traffic.threshold && !traffic.thresholdCrossed {
        traffic.thresholdCrossed = true
        PrintAlert(traffic)
    } else if traffic.hitsTotal < traffic.threshold && traffic.thresholdCrossed {
        traffic.thresholdCrossed = false
        PrintAlert(traffic)
    }
    return traffic
}

func FormatAlert(traffic TrafficAlert) (string) {
    timestamp := int(time.Now().Unix())
    j, _ := json.Marshal(&struct{
        Timestamp         int     `json:"timestamp"`
        MessageType       string  `json:"message_type"`
        AlertType         string  `json:"alert_type"`
        Period            int     `json:"period"`
        Threshold         int     `json:"threshold"`
        CurrentValue      int     `json:"current_value"`
    }{
        Timestamp: timestamp,
        MessageType: "alert",
        AlertType: "traffic_above_threshold",
        Period: 60, // 60 seconds
        Threshold: traffic.threshold,
        CurrentValue: traffic.hitsTotal,
    })
    return string(j)
}

func PrintAlert(traffic TrafficAlert) {
    fmt.Println(FormatAlert(traffic))
}
