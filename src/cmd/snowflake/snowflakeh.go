package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const Epoch int64 = 1413817200000

const ServerBits uint8 = 10
const SequenceBits uint8 = 12

const ServerShift uint8 = SequenceBits
const TimeShift uint8 = SequenceBits + ServerBits

const ServerMax int = -1 ^ (-1 << ServerBits)

const SequenceMask int32 = -1 ^ (-1 << SequenceBits)

type Worker struct {
	serverId      int
	lastTimestamp int64
	sequence      int32
}

type Server struct {
	port    int
	workers chan *Worker
}

func NewWorker(serverId int) *Worker {
	if serverId < 0 || ServerMax < serverId {
		panic(fmt.Errorf("invalid server Id"))
	}
	return &Worker{
		serverId:      serverId,
		lastTimestamp: 0,
		sequence:      0,
	}
}

func NewServer(port, serverId, serverNum int) *Server {
	workers := make(chan *Worker, serverNum)
	for n := 0; n < serverNum; n++ {
		workers <- NewWorker(serverId + n)
	}

	return &Server{
		port:    port,
		workers: workers,
	}
}

func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	c := req.URL.Query().Get("count")
	n := 1
	if c != "" {
		if n, err = strconv.Atoi(c); err != nil || n < 1 || n > 500 {
			http.Error(w, "invalid count; must be a valid integer from 1 to 500", http.StatusBadRequest)
			return
		}
	}

	if ids, err := s.Next(n); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		nl := []byte("\n")
		buf := [32]byte{}
		for _, id := range ids {
			res := strconv.AppendInt(buf[:0], id, 10)
			w.Write(res)
			w.Write(nl)
		}
	}
}

func (s *Server) Next(n int) ([]int64, error) {
	worker := <-s.workers

	ids := make([]int64, n)
	var err error
	for p := 0; p < n; p++ {
		if ids[p], err = worker.Next(); err != nil {
			break
		}
	}

	s.workers <- worker

	return ids, err
}

func (s *Worker) Next() (int64, error) {
	t := now()
	if t < s.lastTimestamp {
		return -1, fmt.Errorf("invalid system clock")
	}
	if t == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & SequenceMask
		if s.sequence == 0 {
			t = s.nextMillis()
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = t
	tp := (t - Epoch) << TimeShift
	sp := int64(s.serverId << ServerShift)
	n := tp | sp | int64(s.sequence)

	return n, nil
}

func (s *Worker) nextMillis() int64 {
	t := now()
	for t <= s.lastTimestamp {
		time.Sleep(100 * time.Microsecond)
		t = now()
	}
	return t
}

func now() int64 {
	return time.Now().UnixNano() / 1000000
}

func main() {
	var portNumber int
	var serverId int
	var serverNum int
	var maxProcs int
	flag.IntVar(&portNumber, "port", 8181, "port")
	flag.IntVar(&serverId, "id", 0, "server id")
	flag.IntVar(&serverNum, "num", 1, "number of workers")
	flag.IntVar(&maxProcs, "proc", 0, "proc")
	flag.Parse()
	if maxProcs == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	server := NewServer(portNumber, serverId, serverNum)
	log.Fatal(server.ListenAndServe())
}
