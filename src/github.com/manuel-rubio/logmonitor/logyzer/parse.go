package logyzer

import (
    "regexp"
    "fmt"
    "strconv"
    "strings"
)

type LogEntry struct {
    clientIP string;
    proxyIPs []string;
    datetime string;
    method string;
    uri string;
    httpver string;
    statuscode int;
    responsetime int;
}

func Parse(line string) (LogEntry)  {
    var re = regexp.MustCompile(`(?m)^(?P<ClientIP>[0-9]+(?:\.[0-9]+){3})(?P<ProxyIPs>(?:, [0-9]+(?:\.[0-9]+){3})*) (\[[^]]+\]) "([A-Z]+) ([^ ]+) ([^"]+)" ([0-9]{3}) ([0-9.]+)$`)
    match := re.FindStringSubmatch(line)
    statuscode, _ := strconv.Atoi(match[7])
    responsetime, _ := strconv.ParseFloat(match[8], 64)
    logEntry := LogEntry{clientIP: match[1],
                         proxyIPs: ParseProxyIPs(match[2]),
                         datetime: match[3],
                         method: match[4],
                         uri: match[5],
                         httpver: match[6],
                         statuscode: statuscode,
                         responsetime: int(responsetime * 1000)}
    return logEntry
}

func ParseProxyIPs(proxyIPs string) []string {
    if proxyIPs == "" {
        var empty []string
        return empty
    }
    return strings.Split(proxyIPs[2:], ", ")
}
