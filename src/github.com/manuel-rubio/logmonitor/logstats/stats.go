package logstats

import (
    "fmt"
    "time"
    "strconv"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

type ProxyHits map[string]int

type Stats struct {
    get int
    post int
    hits int
    forwardedHits int
    proxyUsage ProxyHits
    p95 float64
    badLines int
}

func ProcessStats(entry logyzer.LogEntry, stats Stats) (Stats) {
    if entry.IsBadLine() {
        stats.badLines ++
        return stats
    }
    switch entry.Method() {
    case "GET":
        stats.get += 1
    case "POST":
        stats.post += 1
    default:
        stats.badLines += 1
        return stats
    }
    stats.hits += 1
    stats.forwardedHits += entry.NumProxyIPs()
    for _, proxy := range entry.ProxyIPs() {
        if _, ok := stats.proxyUsage[proxy]; ok {
            stats.proxyUsage[proxy] ++
        } else {
            stats.proxyUsage[proxy] = 1
        }
    }
    // TODO calculate p95
    return stats
}

func StatsLoop(statsChan <-chan logyzer.LogEntry, doneStats <-chan bool) {
    doneTimer := make(chan bool)
    tickTimer := make(chan bool)
    stats := Stats{get: 0,
                   post: 0,
                   hits: 0,
                   forwardedHits: 0,
                   proxyUsage: make(ProxyHits),
                   p95: 0.0,
                   badLines: 0}
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

func FormatStats(stats Stats) (string) {
    timestamp := strconv.Itoa(int(time.Now().Unix()))
    proxy, hits := MostUsedProxy(stats.proxyUsage)
    return `{"timestamp": ` + timestamp + `, "message_type": "stats",` +
           ` "get": ` + strconv.Itoa(stats.get) + `,` +
           ` "post": ` + strconv.Itoa(stats.post) + `,` +
           ` "hits": ` + strconv.Itoa(stats.hits) + `,` +
           ` "forwarded_hits": ` + strconv.Itoa(stats.forwardedHits) + `,` +
           ` "most_used_proxy": "` + proxy + `",` +
           ` "most_used_proxy_hits": ` + strconv.Itoa(hits) + `,` +
           ` "p95": ` + strconv.FormatFloat(stats.p95, 'f', 9, 64) + `,` +
           ` "bad_lines": ` + strconv.Itoa(stats.badLines) + `}`
}

func PrintStats(stats Stats) {
    fmt.Println(FormatStats(stats))
}
