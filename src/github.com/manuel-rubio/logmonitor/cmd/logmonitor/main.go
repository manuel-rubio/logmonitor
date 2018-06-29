package main

import (
    "fmt"
    "os"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

func main() {
    if len(os.Args) != 3 {
        fmt.Println("syntax: logmonitor <access.log> <traffic_threshold>")
        os.Exit(1)
    }
    file := os.Args[1]
    //threshold := os.Args[2]
    lines := make(chan string)
    done := make(chan bool)
    go func() {
        logyzer.Tail(file, lines, done)
    }()
    for {
        select {
        case line := <-lines:
            fmt.Printf("> " + line)
            logyzer.Parse(line)
        case <-done:
            break
        }
    }
}
