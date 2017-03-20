package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type Cushion struct {
	wg sync.WaitGroup

	MessageQueue chan string

	url  string
	port string

	http *http.Server

	mux *http.ServeMux
}

func (this *Cushion) InRequest(w http.ResponseWriter, r *http.Request) {

	message := r.URL.RequestURI() + r.URL.RawQuery

	this.MessageQueue <- message

	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Requests on port " + this.port))

	return

}

func (this *Cushion) CallURL(requestURL string) {

	defer this.wg.Done()

	res, err := http.Get(this.url)

	if err == nil {
		res.Body.Close()
	}

}

func (this *Cushion) OutRequest() {

	for {

		requestURL := <-this.MessageQueue

		this.wg.Add(1)

		go this.CallURL(requestURL)

		this.wg.Wait()

	}
}

func (this *Cushion) QueueSize(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(strconv.Itoa(len(this.MessageQueue)) + " requests in buffer"))

	return

}

func (this *Cushion) Start() {

	this.MessageQueue = make(chan string, 100000)

	this.mux = http.NewServeMux()

	this.mux.HandleFunc("/", this.InRequest)

	this.mux.HandleFunc("/info", this.QueueSize)

	this.http = &http.Server{Addr: this.port, Handler: this.mux}

	go this.OutRequest()

	go this.http.ListenAndServe()

	fmt.Println("Forwarding to ", this.url, " listening on localhost port ", this.port)

}

func main() {

	var wg sync.WaitGroup

	cpus, _ := strconv.Atoi(os.Args[3])

	runtime.GOMAXPROCS(cpus)

	apiservice := Cushion{url: os.Args[1], port: ":" + os.Args[2]}

	wg.Add(1)

	go apiservice.Start()

	wg.Wait()

}
