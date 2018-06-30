package logyzer

import (
    "regexp"
    "strconv"
    "strings"
    "time"
)

type LogEntry struct {
    clientIP string
    proxyIPs []string
    datetime string
    method string
    uri string
    httpver string
    statuscode int
    responsetime time.Duration
    badLine bool
}

func (logentry *LogEntry) ProxyIPs() ([]string) {
    return logentry.proxyIPs
}

func (logentry *LogEntry) NumProxyIPs() (int) {
    return len(logentry.proxyIPs)
}

func (logentry *LogEntry) Method() (string) {
    return logentry.method
}

func (logentry *LogEntry) IsBadLine() (bool) {
    return logentry.badLine
}

func (logentry *LogEntry) ResponseTime() (time.Duration) {
    return logentry.responsetime
}

func Parse(line string) LogEntry  {
    reClientIP := `([0-9]+(?:\.[0-9]+){3})`
    reProxyIPs := `((?:, [0-9]+(?:\.[0-9]+){3})*)`
    reDatetime := `\[([^]]+)\]`
    reQuery := `"([A-Z]+) ([^ ]+) ([^"]+)"`
    reStatus := `([0-9]{3})`
    reReqtime := `([0-9.]+)`
    re := `(?m)^` + reClientIP + reProxyIPs + " " + reDatetime + " " +
          reQuery + " " + reStatus + " " + reReqtime + "$"
    var rg = regexp.MustCompile(re)
    match := rg.FindStringSubmatch(line)
    if match == nil {
        return LogEntry{badLine: true}
    }
    statuscode, _ := strconv.Atoi(match[7])
    responsetime, _ := time.ParseDuration(match[8] + "s")
    logEntry := LogEntry{clientIP: match[1],
                         proxyIPs: ParseProxyIPs(match[2]),
                         datetime: match[3],
                         method: match[4],
                         uri: match[5],
                         httpver: match[6],
                         statuscode: statuscode,
                         responsetime: responsetime,
                         badLine: false}
    return logEntry
}

func ParseProxyIPs(proxyIPs string) []string {
    if proxyIPs == "" {
        var empty []string
        return empty
    }
    return strings.Split(proxyIPs[2:], ", ")
}
