package main

import (
    "fmt"
    "os"
    "strings"
    "math/rand"
    "time"
)

const (
    CLIENT_IP string = "0.0.0.0"
    PROXY1 string = "1.1.1.1"
    PROXY2 string = "2.2.2.2"
    PROXY3 string = "3.3.3.3"
    PROXY4 string = "4.4.4.4"
    PROXY5 string = "5.5.5.5"
    PROXY6 string = "6.6.6.6"
    PROXY7 string = "7.7.7.7"
    PROXY8 string = "8.8.8.8"
    PROXY9 string = "9.9.9.9"
)

func usage() {
    fmt.Println("syntax: genlogs <access.log>")
    os.Exit(1)
}

func main() {
    if len(os.Args) != 2 {
        usage()
    }
    name := os.Args[1]
    file, err := os.OpenFile(name, os.O_WRONLY | os.O_APPEND, os.ModeAppend)
    if err != nil {
        fmt.Errorf("Unable to open file %s: %s", name, err)
        os.Exit(1)
    }
    defer file.Close()

    rand.Seed(time.Now().Unix())

    routes := [][]string{
        // efficient routes
        []string{PROXY1, PROXY2, PROXY3},
        []string{PROXY1, PROXY4, PROXY5},
        []string{PROXY6, PROXY4, PROXY3},
        []string{PROXY6, PROXY2, PROXY7},
        []string{PROXY8, PROXY9, PROXY7},

        // inefficient routes
        []string{PROXY1, PROXY2, PROXY4, PROXY3},
        []string{PROXY2, PROXY4, PROXY5, PROXY3},
        []string{PROXY3, PROXY1, PROXY2, PROXY7},
        []string{PROXY4, PROXY3, PROXY2, PROXY7},
        []string{PROXY5, PROXY1, PROXY2, PROXY3},

        // redefine PROXY4, inefficient PROXY2 and PROXY3
        []string{PROXY2, PROXY3, PROXY4},
        // redefine PROXY9, inefficient PROXY2 and PROXY7
        []string{PROXY2, PROXY7, PROXY9},

        // recheck
        []string{PROXY8, PROXY9, PROXY7},
        []string{PROXY1, PROXY4, PROXY5},
        []string{PROXY6, PROXY4, PROXY3},
    }

    methods := []string{
        "GET",
        "POST",
    }

    for _, route := range routes {
        method := methods[rand.Intn(len(methods))]
        log := genlog(method, route)
        file.WriteString(log)
    }
}

func genlog(method string, route []string) (string) {
    return CLIENT_IP + ", " + strings.Join(route, ", ") + " " +
    "[01/01/1970 00:00:00] \"" + method + " / HTTP/1.1\" 200 " +
    "0.001002003\n"
}
