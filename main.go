package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"sync/atomic"
	"time"
)

var (
	flURL   = flag.String("url", "", "http://pythoninterview.ulaiber.com/")
	flAddrs = flag.String("addrs", "", "host1:8888,host2:8888")
	flPort  = flag.String("port", "8888", "task distribution server")
)

const (
	localhost      = "localhost"
	limitPerSecode = 500
	tokenTimeout   = 30 * time.Second
)

var count, total uint32

func main() {
	flag.Parse()

	setupServer(*flURL != "", ":"+*flPort)

	hosts := strings.Split(*flAddrs, ",")
	hosts = append(hosts, localhost+":"+*flPort)
	addrs := make([]string, 0, len(hosts))
	for i := range hosts {
		if hosts[i] == "" {
			continue
		}
		err := ping(hosts[i])
		if err != nil {
			log.Printf("ping %s error:%s\n", hosts[i], err)
		} else {
			addrs = append(addrs, hosts[i])
			log.Printf("ping %s OK\n", hosts[i])
		}
	}

	if len(addrs) == 0 {
		log.Panic("no server is running")
	}

	key, token, err := getMagic(*flURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), tokenTimeout)
	defer cancel()

	keys := combin(key)
	n := len(addrs)
	ch := make(chan string, n)
	part := len(keys) / n
	if part == 0 {
		part = 1
	}

	for i := 0; i < n; i++ {
		if i == n-1 {
			go postTask(addrs[i], key, *flURL, token, keys[(n-1)*part:], ch)
		} else {
			go postTask(addrs[i], key, *flURL, token, keys[i*part:part*(i+1)], ch)
		}
	}

	select {
	case <-ctx.Done():
		log.Fatalf("运行次数：%d 全排列：%d key=%s token=%s error=%s", atomic.LoadUint32(&total), len(keys), key, token, ctx.Err())
		return

	case v := <-ch:
		log.Printf("运行次数：%d 全排列：%d key=%s token=%s magic=%s", atomic.LoadUint32(&total), len(keys), key, token, v)
		return
	}
}
