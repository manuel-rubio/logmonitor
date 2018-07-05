package logproxy

import(
    "fmt"
    "time"
    "encoding/json"

    "github.com/manuel-rubio/logmonitor/logyzer"
)

const (
    Inefficient = iota
    Efficient
)

type proxyChain []string
type proxyChains []proxyChain

type proxyPtr *Proxy
type proxiesPtr map[string]proxyPtr
type Proxy struct {
    name string
    children proxiesPtr
    parents proxiesPtr
}

func (p Proxy) distance() (int) {
    if len(p.children) > 0 {
        for _, child := range p.children {
            return (*child).distance() + 1
        }
    }
    return 0
}

type proxies map[string]Proxy

func (ps *proxies) processProxyChain(chain proxyChain) {
    childName := chain[len(chain) - 1]
    ps.add(childName)
    for i := len(chain) - 2; i >= 0; i-- {
        parentName := chain[i]
        childName = chain[i + 1]
        ps.addChild(parentName, childName)
    }
}

func (ps *proxies) adjustDistanceUp(name string, proxy *Proxy,
                                    childName string, child *Proxy) {
    if (*proxy).distance() > (*child).distance() {
        proxy.children = make(proxiesPtr)
        proxy.children[childName] = child
        (*ps)[name] = *proxy
        for parentName, parent := range proxy.parents {
            ps.adjustDistanceUp(parentName, parent, name, proxy)
        }
    }
}

func (ps *proxies) add(name string) (Proxy) {
    if proxy, ok := (*ps)[name]; ok {
        if proxy.distance() > 0 {
            for parentName, parent := range proxy.parents {
                ps.adjustDistanceUp(parentName, parent, name, &proxy)
            }
        }
    } else {
        (*ps)[name] = Proxy{name: name,
                            parents: make(proxiesPtr),
                            children: make(proxiesPtr)}
    }
    return (*ps)[name]
}

func (ps *proxies) addChild(parentName string, childName string) {
    if proxy, ok := (*ps)[parentName]; ok {
        if child, ok := (*ps)[childName]; ok {
            proxy_distance := proxy.distance()
            child_distance := child.distance()
            if proxy_distance > (child_distance + 1) {
                for _, ch := range proxy.children {
                    delete(ch.parents, parentName)
                }
                proxy.children = make(proxiesPtr)
            }
            if proxy_distance >= (child_distance + 1) {
                child.parents[parentName] = &proxy
                proxy.children[childName] = &child
                (*ps)[parentName] = proxy
                (*ps)[childName] = child
            }
        } else {
            child = Proxy{name: childName,
                          children: make(proxiesPtr),
                          parents: make(proxiesPtr)}
            proxy_distance := proxy.distance()
            if proxy_distance > 1 {
                for _, ch := range proxy.children {
                    delete(ch.parents, parentName)
                }
                proxy.children = make(proxiesPtr)
            }
            if proxy_distance >= 1 {
                child.parents[parentName] = &proxy
                proxy.children[childName] = &child
                (*ps)[parentName] = proxy
                (*ps)[childName] = child
            }
        }
    } else {
        if child, ok := (*ps)[childName]; ok {
            proxy = Proxy{name: parentName,
                          children: make(proxiesPtr),
                          parents: make(proxiesPtr)}
            proxy.children[childName] = &child
            child.parents[parentName] = &proxy
            (*ps)[parentName] = proxy
            (*ps)[childName] = child
        } else {
            panic(fmt.Sprintf("node %d and %d doesn't exist... this mustn't happen!\n",
                              parentName, childName))
        }
    }
}

func (ps *proxies) getDistance(name string) (int) {
    return (*ps)[name].distance()
}

func (proxies *proxies) adjustPaths(pc proxyChain, i int,
                                    ip *inefficientProxies) (int) {
    pname := pc[0]
    inDistance := len(pc) - 1
    distance := proxies.getDistance(pname)
    inefficient := -1
    if distance < inDistance {
        // inefficient
        ip.proxies = append(ip.proxies, pname)
        inefficient = i
    }
    if len(pc) > 1 {
        j := proxies.adjustPaths(pc[1:], i + 1, ip)
        if inefficient < 0 {
            inefficient = j
        }
    }
    return inefficient
}

func (ps *proxies) generateChains(prefix proxyChain, name string) (proxyChains) {
    var chains proxyChains
    prefix = append(prefix, name)
    children := (*ps)[name].children
    if len(children) > 0 {
        for proxyName, _ := range children {
            distance := ps.getDistance(name)
            if distance > 0 {
                genChains := ps.generateChains(prefix, proxyName)
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

type inefficientProxies struct {
    originalChain proxyChain
    proxies proxyChain
    efficientChains proxyChains
}

func ProxyLoop(logs <-chan logyzer.LogEntry, done <-chan bool) {
    proxies := make(proxies)
    for {
        select {
        case entry := <-logs:
            ProcessLog(entry, &proxies)
        case <-done:
            return
        }
    }
}

func ProcessLog(entry logyzer.LogEntry, proxies *proxies) {
    if entry.IsBadLine() {
        return
    }
    inProxies := entry.ProxyIPs()
    ip := inefficientProxies{originalChain: inProxies}
    proxies.processProxyChain(inProxies)
    inefficient := proxies.adjustPaths(inProxies, 0, &ip)
    if inefficient >= 0 {
        prefix := make([]string, len(inProxies[:inefficient]))
        copy(prefix, inProxies[:inefficient])
        name := inProxies[inefficient]
        ip.efficientChains = proxies.generateChains(prefix, name)
        PrintAlert(&ip)
    }
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
