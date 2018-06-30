package logstats

import (
    "fmt"
    "time"

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

func PrintStats(stats Stats) {
    // TODO generate output
    fmt.Println(stats)
}
