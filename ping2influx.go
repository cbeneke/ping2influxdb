package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "time"

    "github.com/influxdata/influxdb/client/v2"
    "github.com/tatsushid/go-fastping"
)

const (
    database = "ping"
    host = "http://127.0.0.1:8086"
    username = ""
    password = ""
)

var (
    pinger *fastping.Pinger
    hostname string
    remoteaddr *net.IPAddr
    influxdb client.Client
)

func init() {
    hostname = string(os.Args[1])
    pinger = fastping.NewPinger()

    remoteaddr, err := net.ResolveIPAddr("ip4:icmp", hostname)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    pinger.OnRecv = write
    pinger.AddIPAddr(remoteaddr)

    influxdb, err = client.NewHTTPClient(client.HTTPConfig{
        Addr:     host,
        Username: username,
        Password: password,
    })
}

func ping() {
    err := pinger.Run()
    if err != nil {
        fmt.Println(err)
    }
}

func write(addr *net.IPAddr, rtt time.Duration) {
    fmt.Printf("ping to %s : %v\n", addr.String(), rtt)

    // Create a new point batch
    bp, err := client.NewBatchPoints(client.BatchPointsConfig{
        Database:  database,
        Precision: "s",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a point and add to batch
    cur_host, err := os.Hostname()
    tags := map [string] string {
        "host": cur_host,
    }

    fields := map [string] interface{} {
        "rtt": float64(rtt.Nanoseconds()) / 1000000,
        "address": addr.String(),
        "hostname": hostname,
    }

    pt, err := client.NewPoint(hostname, tags, fields, time.Now())
    if err != nil {
        log.Fatal(err)
    }
    bp.AddPoint(pt)

    // Write the batch
    if err := influxdb.Write(bp); err != nil {
        log.Fatal(err)
    }
}

func main() {
    for {
        ping()
    }
}
