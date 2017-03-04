package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// ensureBodyClose close *http.Response
func ensureBodyClose(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.CopyN(ioutil.Discard, resp.Body, 512)

		resp.Body.Close()
	}
}

func ping(addr string) error {
	resp, err := http.Get("http://" + addr + "/_ping")
	if err != nil {
		return err
	}
	defer ensureBodyClose(resp)

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.Errorf("ping %s StatusCode:%d", addr, resp.StatusCode)
}

func postTask(addr, origin, url, token string, keys []string, ch chan<- string) error {
	log.Println("post task to:", addr)

	task := taskRequest{
		URL:    url,
		Origin: origin,
		Token:  token,
		Keys:   keys,
	}

	buffer := bytes.NewBuffer(nil)

	err := json.NewEncoder(buffer).Encode(task)
	if err != nil {
		log.Printf("addr=%s", addr)
		return err
	}

	resp, err := http.Post("http://"+addr+"/task", "application/json", buffer)
	if err != nil {
		return err
	}
	defer ensureBodyClose(resp)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		ch <- string(data)
		return nil
	}

	return errors.Errorf("addr:%s StatusCode=%d,body=%s", addr, resp.StatusCode, data)
}

type taskRequest struct {
	URL    string
	Origin string
	Token  string
	Keys   []string
}

func runTasks(w http.ResponseWriter, r *http.Request) {
	log.Println("task from ", r.RemoteAddr)

	var req = taskRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Fprint(w, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), tokenTimeout)
	defer cancel()

	keys := req.Keys
	n := 30
	ch := make(chan string, n)
	part := len(keys) / n
	if part == 0 {
		part = 1
	} else {
		go checkAndCount(ctx, keys[(n-1)*part:], req.URL, req.Token, ch)
	}

	for i := 0; i < n-1; i++ {
		go checkAndCount(ctx, keys[i*part:part*(i+1)], req.URL, req.Token, ch)
	}

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Printf("运行次数：%d 全排列：%d key=%s token=%s error=%s", atomic.LoadUint32(&total), len(keys), req.Origin, req.Token, err)

		if err != nil {
			fmt.Fprint(w, err.Error())
		}
		w.WriteHeader(http.StatusInternalServerError)

	case v := <-ch:
		log.Printf("运行次数：%d 全排列：%d key=%s token=%s magic=%s", atomic.LoadUint32(&total), len(keys), req.Origin, req.Token, v)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(v))
	}
}

func checkAndCount(ctx context.Context, keys []string, url, token string, ch chan<- string) {
	last := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := range keys {

		if atomic.CompareAndSwapUint32(&count, limitPerSecode, limitPerSecode) {
			if t := time.Second - time.Now().Sub(last); t > 0 {
				log.Println("sleep:", t.String())

				time.Sleep(t)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			atomic.StoreUint32(&count, 0)
			last = time.Now()
		default:
		}

		atomic.AddUint32(&count, 1)
		atomic.AddUint32(&total, 1)

		err := checkResult(url, keys[i], token)
		if err == nil {
			ch <- keys[i]
			return
		} else {
			log.Println(keys[i], err)
		}
	}
}

// setup server,if background is true,run server in a new goroutine.
func setupServer(background bool, addr string) {
	mux := http.NewServeMux()

	// handlers
	mux.HandleFunc("/task", runTasks)
	mux.HandleFunc("/_ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if background {
		go func() {
			log.Println("setup Server in background", addr)
			log.Fatal(http.ListenAndServe(addr, mux))
		}()

	} else {
		log.Println("setup Server", addr)

		log.Fatal(http.ListenAndServe(addr, mux))
	}
}
