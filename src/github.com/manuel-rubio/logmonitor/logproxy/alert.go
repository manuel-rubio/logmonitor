package logproxy

import(
    "fmt"
    "time"
    "encoding/json"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

type proxyChain []string
type proxyChains []proxyChain

type proxyPath map[string]bool

type proxy struct {
    paths proxyPath
    distance int
}

type efficientProxies map[string]proxy

func (ep *efficientProxies) process(pc proxyChain, i int,
                                    ie *inefficientProxies) (string, int) {
    pname := pc[0]
    var path proxyChain
    if len(pc) > 1 {
        path = pc[1:]
    }
    if p, ok := (*ep)[pname]; ok {
        if len(path) > 0 {
            n, d := ep.process(path, i + 1, ie)
            if d < p.distance {
                p.distance = d
                p.paths = make(proxyPath)
                p.paths[n] = true
                (*ep)[pname] = p
            } else if d > p.distance {
                prefix := make(proxyChain, i)
                copy(prefix, ie.originalChain[:i + 1])
                ep.triggerAlert(prefix, pname, ie)
                return pname, p.distance + 1
            } else {
                p.paths[n] = true
            }
            return pname, d + 1
        } else {
            if 0 < p.distance {
                p.distance = 0
                p.paths = make(proxyPath)
                (*ep)[pname] = p
            }
            return pname, 1
        }
    } else {
        d := 0
        h := make(proxyPath)
        if len(pc) > 1 {
            var n string
            n, d = ep.process(pc[1:], i + 1, ie)
            h[n] = true
        }
        (*ep)[pname] = proxy{paths: h, distance: d}
        return pname, d + 1
    }
}

func (ep *efficientProxies) generateChains(prefix proxyChain, name string) (proxyChains) {
    var chains proxyChains
    prefix = append(prefix, name)
    paths := (*ep)[name].paths
    if len(paths) > 0 {
        for proxy, _ := range paths {
            if (*ep)[name].distance > 0 {
                genChains := ep.generateChains(prefix, proxy)
                chains = append(chains, genChains...)
            } else {
                chains = append(chains, prefix)
            }
        }
    } else {
        chains = append(chains, prefix)
    }
    return chains
}

func (ep *efficientProxies) triggerAlert(prefix proxyChain, name string,
                                         ie *inefficientProxies) {
    ie.proxies = append(ie.proxies, name)
    ie.efficientChains = ep.generateChains(prefix, name)
    PrintAlert(ie)
}

type inefficientProxies struct {
    originalChain proxyChain
    proxies proxyChain
    efficientChains proxyChains
}

func ProxyLoop(logs <-chan logyzer.LogEntry, done <-chan bool) {
    proxies := make(efficientProxies)
    for {
        select {
        case entry := <-logs:
            ProcessLog(entry, &proxies)
        case <-done:
            return
        }
    }
}

func ProcessLog(entry logyzer.LogEntry, proxies *efficientProxies) {
    if entry.IsBadLine() {
        return
    }
    inProxies := entry.ProxyIPs()
    ip := inefficientProxies{originalChain: inProxies}
    proxies.process(inProxies, 0, &ip)
}

func FormatAlert(ip *inefficientProxies) (string) {
    timestamp := int(time.Now().Unix())
    j, _ := json.Marshal(&struct{
        Timestamp            int         `json:"timestamp"`
        MessageType          string      `json:"message_type"`
        AlertType            string      `json:"alert_type"`
        ProxyChain           proxyChain  `json:"proxy_chain"`
        InefficientAddresses proxyChain  `json:"inefficient_addresses"`
        EfficientProxyChains proxyChains `json:"efficient_proxy_chains"`
    }{
        Timestamp: timestamp,
        MessageType: "alert",
        AlertType: "inefficient_proxy_chain",
        ProxyChain: ip.originalChain,
        InefficientAddresses: ip.proxies,
        EfficientProxyChains: ip.efficientChains,
    })
    return string(j)
}

func PrintAlert(ip *inefficientProxies) {
    fmt.Println(FormatAlert(ip))
}
