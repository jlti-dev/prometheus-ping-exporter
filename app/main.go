package main

import (
        "fmt"
        "net"
        "net/http"
        "os"
        "os/signal"
        "syscall"
        "time"
        "strconv"
        "sync"
        "github.com/tatsushid/go-fastping"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "github.com/prometheus/client_golang/prometheus/promauto"
)

type response struct {
        addr *net.IPAddr
        rtt  time.Duration
}
type metric struct {
        cnt int64
        err int64
        hig int64
        low int64
        avg int64
        last int64
        pkt int64
}
//COUNTER:
var metric_last_pkt = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "scrape_packets",
        Help: "Number of Ping attempts since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_pkt = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "total_packets",
        Help: "Number of Ping attempts",
        },
        []string{"host", "name", "group"})
var metric_last_suc = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "scrape_success",
        Help: "Number of successful Ping attempts since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_suc = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "total_success",
        Help: "Number of successful Ping attempts",
        },
        []string{"host", "name", "group"})
var metric_last_err = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "scrape_fail",
        Help: "Number of failed Ping attempts since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_err = promauto.NewCounterVec(prometheus.CounterOpts{
        Namespace: "ping",
        Name: "total_fail",
        Help: "Number of failed Ping attempts",
        },
        []string{"host", "name", "group"})

//GAUGE
var metric_last_hig = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "scrape_high",
        Help: "Highest Ping since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_hig = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "total_high",
        Help: "Highest Ping since start of exporter",
        },
        []string{"host", "name", "group"})
var metric_last_low = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "scrape_low",
        Help: "Lowest Ping since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_low = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "total_low",
        Help: "Lowest Ping since start of exporter",
        },
        []string{"host", "name", "group"})
var metric_last_avg = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "scrape_avg",
        Help: "Average Ping since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_avg = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "total_avg",
        Help: "Average Ping since start of exporter",
        },
        []string{"host", "name", "group"})
var metric_last_last = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "scrape_last",
        Help: "Last Ping since last scrape",
        },
        []string{"host", "name", "group"})
var metric_total_last = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "ping",
        Name: "total_last",
        Help: "Last Ping since start of exporter",
        },
        []string{"host", "name", "group"})

var mutex = &sync.Mutex{}
var metric_ever = make(map[string]*metric)
var metric_last = make(map[string]*metric)
var results = make(map[string]*response)
var labelsName = make(map[string]string)
var labelsGroup = make(map[string]string)

func update_metric(metric *metric, rtt time.Duration) *metric {
        m := metric
        m.last = rtt.Microseconds()
        if m.hig < m.last { m.hig = m.last }
        if m.low == 0 || m.low < m.last { m.low = m.last }
        m.avg = (m.avg * m.cnt + m.last) / (m.cnt + 1)
        m.cnt = m.cnt + 1
        m.pkt = m.pkt + 1
        //fmt.Printf("last: %d, hig: %d, low: %d, avg: %d, cnt: %d\n", m.last, m                                                                                                                                                                                                                                             .hig, m.low, m.avg, m.cnt)
        return m
}

func handleRequest(h http.Handler) http.Handler{
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
                h.ServeHTTP(w,r)
                resetMetrics()
        })
}
func resetMetrics(){
        //clearen der metrics
        metric_last_pkt.Reset()
        metric_last_suc.Reset()
        metric_last_err.Reset()
        mutex.Lock()
        defer mutex.Unlock()
        for host, _ := range metric_last {
                metric_last[host] = nil
        }

}
func publishMetrics(host string){
        me := metric_ever[host]
        ml := metric_last[host]

        divisor := float64(1000)
        value := float64(me.hig) / divisor
        metric_total_hig.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Set(value)
        value = float64(ml.hig) / divisor
        metric_last_hig.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Set(value)

        value = float64(me.low) / divisor
        metric_total_low.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Set(value)
        value = float64(ml.low) / divisor
        metric_last_low.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Set(value)

        value = float64(me.avg) / divisor
        metric_total_avg.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Set(value)
        value = float64(ml.avg) / divisor
        metric_last_avg.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Set(value)

        value = float64(me.last) / divisor
        metric_total_last.WithLabelValues(host,labelsName[host],labelsGroup[host                                                                                                                                                                                                                                             ]).Set(value)
        value = float64(ml.last) / divisor
        metric_last_last.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Set(value)

}
func addMetricError(host string){
        metric_total_err.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Inc()
        metric_last_err.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Inc()

        metric_total_pkt.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Inc()
        metric_last_pkt.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Inc()
}
func addMetricSuccess(host string){
        metric_total_suc.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Inc()
        metric_last_suc.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Inc()

        metric_total_pkt.WithLabelValues(host,labelsName[host],labelsGroup[host]                                                                                                                                                                                                                                             ).Inc()
        metric_last_pkt.WithLabelValues(host,labelsName[host],labelsGroup[host])                                                                                                                                                                                                                                             .Inc()
}
func doWork(){
        mutex.Lock()
        defer mutex.Unlock()
        for host, r := range results {
                me := metric_ever[host]
                ml := metric_last[host]
                if(me == nil){
                        me = &metric{}
                }
                if(ml == nil){
                        ml = &metric{}
                }

                if r == nil {
                        me.pkt ++
                        me.err ++
                        metric_ever[host] = me

                        ml.pkt ++
                        ml.err ++
                        metric_last[host] = ml
                        addMetricError(host)
                        fmt.Printf("[%s] : unreachable\n", host)
                } else {
                        metric_ever[host] = update_metric(me,r.rtt)
                        metric_last[host] = update_metric(ml,r.rtt)
                        addMetricSuccess(host)
                        fmt.Printf("[%s] avg: %d, last: %d\n", host, metric_ever                                                                                                                                                                                                                                             [host].avg, metric_ever[host].last)
                        //fmt.Printf("%s : %v %v\n", host, r.rtt, time.Now())
                }
                results[host] = nil
                publishMetrics(host)
        }

}
func serve(){
        http.Handle("/metrics", handleRequest(promhttp.Handler()))
        http.ListenAndServe(":8080", nil)
}
func loadEnv(p *fastping.Pinger) {
        i := 0
        for{
                i ++
                env := os.Getenv("IP_" + strconv.Itoa(i))
                if(len(env) == 0) {break}
                p.AddIP(env)
                results[env] = nil
                metric_ever[env] = nil
                metric_last[env] = nil
                fmt.Println("Ping to Server: " + env)

                name := os.Getenv("NAME_" + strconv.Itoa(i))
                group := os.Getenv("GROUP_" + strconv.Itoa(i))
                if(len(name) == 0){
                        labelsName[env] = env
                }else{
                        labelsName[env] = name
                }
                if(len(group) == 0){
                        labelsGroup[env] = "default"
                }else{
                        labelsGroup[env] = group
                }
        }
}

func main() {
        p := fastping.NewPinger()

        loadEnv(p)
        onRecv, onIdle := make(chan *response), make(chan bool)
        p.OnRecv = func(addr *net.IPAddr, t time.Duration) {
                onRecv <- &response{addr: addr, rtt: t}
        }
        p.OnIdle = func() {
                onIdle <- true
        }

        p.MaxRTT = time.Second
        p.RunLoop()

        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        signal.Notify(c, syscall.SIGTERM)
        go serve()
loop:
        for {
                select {
                case <-c:
                        fmt.Println("get interrupted")
                        break loop
                case res := <-onRecv:
                        if _, ok := results[res.addr.String()]; ok {

                                results[res.addr.String()] = res
                        }
                case <-onIdle:
                        doWork()
                case <-p.Done():
                        if err := p.Err(); err != nil {
                                fmt.Println("Ping failed:", err)
                        }
                        break loop
                }
        }
        signal.Stop(c)
        p.Stop()
}
