package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync"
    "strconv"

    "github.com/manuel-rubio/logmonitor/logyzer"
    "github.com/manuel-rubio/logmonitor/logstats"
    "github.com/manuel-rubio/logmonitor/logtraffic"
    "github.com/manuel-rubio/logmonitor/logproxy"
)

func usage() {
    fmt.Println("syntax: logmonitor <access.log> <traffic_threshold>")
    os.Exit(1)
}

func main() {
    if len(os.Args) != 3 {
        usage()
    }
    var wg sync.WaitGroup
    file := os.Args[1]
    threshold, err := strconv.Atoi(os.Args[2])
    if err != nil {
        usage()
    }

    // Process and channels to handle Tail (file read)
    lines := make(chan string)
    doneTail := make(chan bool)
    quitTail := make(chan bool)
    run(&wg, func() {
        logyzer.Tail(file, lines, doneTail, quitTail)
    })

    // Process and channels to handle Stats
    stats := make(chan logyzer.LogEntry)
    doneStats := make(chan bool)
    run(&wg, func() {
        logstats.StatsLoop(stats, doneStats)
    })

    // Process and channels to handle Traffic Alert
    trafficAlert := make(chan logyzer.LogEntry)
    doneTraffic := make(chan bool)
    run(&wg, func() {
        logtraffic.TrafficLoop(trafficAlert, doneTraffic, threshold)
    })

    // Process and channels to handle Proxy Alert
    proxyAlert := make(chan logyzer.LogEntry)
    doneProxy := make(chan bool)
    run(&wg, func() {
        logproxy.ProxyLoop(proxyAlert, doneProxy)
    })

    // Process to handle Ctrl+C
    run(&wg, func() {
        HandleBreak([]chan<- bool{doneStats, doneTraffic, doneProxy, quitTail})
    })

    alerts := []chan<- logyzer.LogEntry{stats, trafficAlert, proxyAlert}
    MainLoop(lines, alerts, doneTail)
    wg.Wait()
}

func run(wg *sync.WaitGroup, code func()) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        code()
    }()
}

func HandleBreak(done []chan<- bool) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <- sigChan
    for _, ch := range done {
        ch <- true
    }
}

// MainLoop is in charge to read lines incoming from Tail process (reading file)
// and drive them to the Stats and Alerts processes.
func MainLoop(lines <-chan string, hooks []chan<- logyzer.LogEntry,
              done <-chan bool) {
    for {
        select {
        case line := <-lines:
            parsed := logyzer.Parse(line)
            for _, ch := range hooks {
                ch <- parsed
            }
        case <-done:
            return
        }
    }
}
