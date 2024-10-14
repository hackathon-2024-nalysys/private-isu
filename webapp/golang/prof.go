package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Profile struct {
	LastStart int64
	Time      int64
	Count     int64
}

// ここにためる
var profiles = map[string]*Profile{}

var ctr atomic.Int64
var times sync.Map

type Report struct {
	Time int64
	Name string
}

var report = make(chan Report, 100)

// はかりはじめる
func startTime() int64 {
	id := ctr.Add(1)
	times.Store(id, time.Now())
	return id
}

// はかりおわる
func endTime(name string, id int64) {
	startTimeAny, _ := times.Load(id)
	duration := time.Since(startTimeAny.(time.Time))
	report <- Report{duration.Nanoseconds(), name}
}

func dumpProfiles() {
	println("---DUMP PROFILE---")
	// convert to per count time and name pair
	ptpair := []struct {
		Name string
		Time int64
	}{}
	for name, p := range profiles {
		ptpair = append(ptpair, struct {
			Name string
			Time int64
		}{name, p.Time / p.Count})
	}
	// sort by time using sort.Slice
	sort.Slice(ptpair, func(i, j int) bool {
		return ptpair[i].Time > ptpair[j].Time
	})

	// dump ptpair with profiles
	for _, pt := range ptpair {
		p := profiles[pt.Name]
		perTime := time.Duration(pt.Time).String()
		totalTime := time.Duration(p.Time).String()
		println(pt.Name, perTime, totalTime, p.Count)
	}
}

func registerProfSignalHandler() {
	go func() {
		r := <-report
		p := profiles[r.Name]
		if p == nil {
			p = &Profile{}
			profiles[r.Name] = p
		}
		p.Count++
		p.Time += r.Time
	}()
	// dump on USR1
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR1)
		for {
			<-c
			stopProfile()
			dumpProfiles()
		}
	}()
	// clear on USR2
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR2)
		for {
			<-c
			profiles = map[string]*Profile{}
			startProfile()
		}
	}()
}

// contextに値をセットするmiddlewareの例
func ProfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := startTime()
		next.ServeHTTP(w, r)
		endTime(r.URL.RequestURI(), id)
	})
}

func startProfile() error {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		return err
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}
	return nil
}

func stopProfile() {
	pprof.StopCPUProfile()
}
