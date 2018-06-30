package logstats

import (
    "fmt"
    "time"
    "encoding/json"
    "math"
    "sort"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

const respTimeSampleSize int = 1000

type ProxyHits map[string]int

type timeSlice []time.Duration

type ResposeTimes struct {
    times timeSlice
    count int
}

func (r *ResposeTimes) AddTime(t time.Duration) {
    r.count ++
    r.times[(r.count - 1) % respTimeSampleSize] = t
}

// Sort Interface: Len
func (ts timeSlice) Len() (int) {
    return len(ts)
}

// Sort Interface: Less
func (ts timeSlice) Less(i, j int) (bool) {
    return int64(ts[i]) < int64(ts[j])
}

// Sort Interface: Swap
func (ts timeSlice) Swap(i, j int) {
    ts[i], ts[j] = ts[j], ts[i]
}

type Stats struct {
    get int
    post int
    hits int
    forwardedHits int
    proxyUsage ProxyHits
    responseTimes ResposeTimes
    badLines int
}

func ProcessStats(entry logyzer.LogEntry, stats Stats) (Stats) {
    if entry.IsBadLine() {
        stats.badLines ++
        return stats
    }
    switch entry.Method() {
    case "GET":
        stats.get ++
    case "POST":
        stats.post ++
    default:
        stats.badLines ++
        return stats
    }
    stats.hits ++
    stats.forwardedHits += entry.NumProxyIPs()
    for _, proxy := range entry.ProxyIPs() {
        if _, ok := stats.proxyUsage[proxy]; ok {
            stats.proxyUsage[proxy] ++
        } else {
            stats.proxyUsage[proxy] = 1
        }
    }
    stats.responseTimes.AddTime(entry.ResponseTime())
    return stats
}

func StatsLoop(statsChan <-chan logyzer.LogEntry, doneStats <-chan bool) {
    doneTimer := make(chan bool)
    tickTimer := make(chan bool)
    stats := Stats{
        get: 0,
        post: 0,
        hits: 0,
        forwardedHits: 0,
        proxyUsage: make(ProxyHits),
        responseTimes: ResposeTimes{
            times: make(timeSlice, respTimeSampleSize),
            count: 0,
        },
        badLines: 0,
    }
    go Timer(tickTimer, doneTimer)
    for {
        select {
        case entry := <-statsChan:
            stats = ProcessStats(entry, stats)
        case <-tickTimer:
            PrintStats(stats)
        case <-doneStats:
            doneTimer <- true
            PrintStats(stats)
            return
        }
    }
}

func Timer(tick chan<- bool, doneTimer <-chan bool) {
    for {
        select {
        case <-time.After(10 * time.Second):
            tick <- true
        case <-doneTimer:
            return
        }
    }
}

func MostUsedProxy(proxies ProxyHits) (string, int) {
    if len(proxies) == 0 {
        return "", 0
    }
    var proxy string
    proxyHits := 0
    for p, hits := range proxies {
        if hits > proxyHits {
            proxyHits = hits
            proxy = p
        }
    }
    return proxy, proxyHits
}

func P95(responseTimes ResposeTimes) (float64) {
    if responseTimes.count == 0 {
        return 0.0
    }
    count := float64(responseTimes.count)
    samples := int(math.Min(count, float64(respTimeSampleSize)))
    times := make(timeSlice, samples)
    copy(times, responseTimes.times[:samples])
    sort.Sort(times)
    return float64(times[int(count * 0.95 + 0.5) - 1]) / float64(time.Second)
}

func FormatStats(stats Stats) (string) {
    timestamp := int(time.Now().Unix())
    proxy, hits := MostUsedProxy(stats.proxyUsage)
    p95 := P95(stats.responseTimes)
    j, _ := json.Marshal(&struct{
        Timestamp         int     `json:"timestamp"`
        MessageType       string  `json:"message_type"`
        Get               int     `json:"get"`
        Post              int     `json:"post"`
        Hits              int     `json:"hits"`
        ForwardedHits     int     `json:"forwarded_hits"`
        MostUsedProxy     string  `json:"most_used_proxy"`
        MostUsedProxyHits int     `json:"most_used_proxy_hits"`
        P95               float64 `json:"p95"`
        BadLines          int     `json:"bad_lines"`
    }{
        Timestamp: timestamp,
        MessageType: "stats",
        Get: stats.get,
        Post: stats.post,
        Hits: stats.hits,
        ForwardedHits: stats.forwardedHits,
        MostUsedProxy: proxy,
        MostUsedProxyHits: hits,
        P95: p95,
        BadLines: stats.badLines,
    })
    return string(j)
}

func PrintStats(stats Stats) {
    fmt.Println(FormatStats(stats))
}
