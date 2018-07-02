package logproxy

import(
    "fmt"
    "time"
    "encoding/json"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

type proxyChain []string

func (pc proxyChain) equal(other proxyChain) (bool) {
    for i, v := range pc {
        if v != other[i] {
            return false
        }
    }
    return true
}

type proxyChains []proxyChain
type efficientProxies map[string]proxyChains

type inefficientProxies struct {
    proxies []string
    efficientChains proxyChains
}

func (ip *inefficientProxies) add(proxy string, chains proxyChains) {
    addProxy := false
    for _, chain1 := range chains {
        isEqual := false
        for _, chain2 := range ip.efficientChains {
            if chain1.equal(chain2) {
                isEqual = true
                break
            }
        }
        if !isEqual {
            ip.efficientChains = append(ip.efficientChains, chain1)
            addProxy = true
        }
    }
    if addProxy {
        ip.proxies = append(ip.proxies, proxy)
    }
}

func (ep efficientProxies) isProxy(proxy string) (bool) {
    if _, ok := ep[proxy]; ok {
        return true
    }
    return false
}

func (ep *efficientProxies) set(proxy string, proxyPath proxyChain) {
    (*ep)[proxy] = proxyChains{proxyPath}
}

func (ep *efficientProxies) add(proxy string, proxyPath proxyChain) {
    (*ep)[proxy] = append((*ep)[proxy], proxyPath)
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
    var ip inefficientProxies
    for i := len(inProxies) - 1; i >= 0; i-- {
        proxy := inProxies[i]
        var proxyPath proxyChain
        if len(inProxies) > i {
            proxyPath = inProxies[(i+1):]
        }
        if proxies.isProxy(proxy) {
            Check(proxies, &ip, inProxies[:i+1], proxy, proxyPath)
        } else {
            proxies.set(proxy, proxyPath)
        }
    }
    if len(ip.proxies) > 0 {
        PrintAlert(inProxies, ip)
    }
    return
}

func Check(ep *efficientProxies, ip *inefficientProxies,
           prefixPath proxyChain, proxy string, proxyPath proxyChain) {
    if len((*ep)[proxy][0]) > len(proxyPath) {
        ep.set(proxy, proxyPath)
    } else if len((*ep)[proxy][0]) == len(proxyPath) {
        for _, pp := range (*ep)[proxy] {
            if pp.equal(proxyPath) {
                return
            }
        }
        ep.add(proxy, proxyPath)
    } else {
        var options proxyChains
        for _, chain := range (*ep)[proxy] {
            option := make(proxyChain, len(prefixPath))
            copy(option, prefixPath)
            option = append(option, chain...)
            options = append(options, option)
        }
        ip.add(proxy, options)
    }
}

func FormatAlert(logProxyChain proxyChain,
                 ip inefficientProxies) (string) {
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
        ProxyChain: logProxyChain,
        InefficientAddresses: ip.proxies,
        EfficientProxyChains: ip.efficientChains,
    })
    return string(j)
}

func PrintAlert(logProxyChain proxyChain, ip inefficientProxies) {
    fmt.Println(FormatAlert(logProxyChain, ip))
}
