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

	//how much request buffering required
	buffer int

	cpus int

	http *http.Server

	mux *http.ServeMux
}

func (this *Cushion) InRequest(w http.ResponseWriter, r *http.Request) {

	message := r.URL.RequestURI()

	this.MessageQueue <- message

	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Requests on port " + this.port))

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

		for loop := 0; loop < this.cpus; loop++ {

			requestURL := <-this.MessageQueue

			this.wg.Add(1)

			go this.CallURL(requestURL)

		}

		this.wg.Wait()
	}
}

func (this *Cushion) QueueSize(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(strconv.Itoa(len(this.MessageQueue)) + " requests in buffer"))

}

func (this *Cushion) Start(cpus int, buffer int) {

	this.cpus = cpus

	this.buffer = buffer

	this.MessageQueue = make(chan string, this.buffer)

	this.mux = http.NewServeMux()

	this.mux.HandleFunc("/", this.InRequest)

	this.mux.HandleFunc("/info", this.QueueSize)

	this.http = &http.Server{Addr: this.port, Handler: this.mux}

	go this.OutRequest()

	go this.http.ListenAndServe()

	fmt.Println("Forwarding to ", this.url, " listening on localhost port ", this.port)

}

func main() {

	if len(os.Args) < 5 {

		fmt.Println("Please pass Redirect URL Listener_Port Number_of_threads Buffer_Size")
		os.Exit(0)
	}

	cpus, _ := strconv.Atoi(os.Args[3])

	buffer, _ := strconv.Atoi(os.Args[4])

	runtime.GOMAXPROCS(cpus)

	//you can create mote objects like this in the next line and call its Start() function
	apiservice := Cushion{url: os.Args[1], port: ":" + os.Args[2]}
	apiservice.Start(cpus, buffer)

	//This should be the last. All to be before this. This line is non-ending loop
	select {}

}
