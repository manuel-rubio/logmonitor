package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync"

    "github.com/manuel-rubio/logmonitor/logyzer"
    "github.com/manuel-rubio/logmonitor/logstats"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Println("syntax: logmonitor <access.log> <traffic_threshold>")
        os.Exit(1)
    }
    var wg sync.WaitGroup
    file := os.Args[1]
    //threshold := os.Args[2]
    lines := make(chan string)
    doneTail := make(chan bool)
    quitTail := make(chan bool)
    go func() {
        wg.Add(1)
        logyzer.Tail(file, lines, doneTail, quitTail)
        wg.Done()
    }()

    stats := make(chan logyzer.LogEntry)
    doneStats := make(chan bool)
    go func() {
        wg.Add(1)
        logstats.StatsLoop(stats, doneStats)
        wg.Done()
    }()

    go func() {
        wg.Add(1)
        HandleBreak(doneStats, quitTail)
        wg.Done()
    }()

    MainLoop(lines, stats, doneTail)
    wg.Wait()
}

func HandleBreak(doneStats chan bool, quitTail chan bool) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <- sigChan
    quitTail <- true
    doneStats <- true
}

func MainLoop(lines chan string, stats chan logyzer.LogEntry, done chan bool) {
    for {
        select {
        case line := <-lines:
            stats <- logyzer.Parse(line)
        case <-done:
            return
        }
    }
}
