package main

import (
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
)

type Profile struct {
	LastStart int64
	Time      int64
	Count     int64
}

// ここにためる
// var profiles = map[string]*Profile{}
var profiles = sync.Map{}

// はかりはじめる
func startTime(name string) {
	pany, ok := profiles.Load(name)
	var p *Profile
	if !ok {
		p = &Profile{}
		profiles.Store(name, p)
	} else {
		p = pany.(*Profile)
	}

	// println("start", name)
	// get current time
	p.LastStart = time.Now().UnixNano()
}

// はかりおわる
func endTime(name string) {
	pany, ok := profiles.Load(name)
	var p *Profile
	if !ok {
		p = &Profile{}
		profiles.Store(name, p)
	} else {
		p = pany.(*Profile)
	}

	// get current time
	p.Time += time.Now().UnixNano() - p.LastStart
	p.Count++
}

func dumpProfiles() {
	println("---DUMP PROFILE---")
	// convert to per count time and name pair
	ptpair := []struct {
		Name string
		Time int64
	}{}
	profiles.Range(func(name any, pany any) bool {
		p := pany.(*Profile)
		nameStr := name.(string)
		ptpair = append(ptpair, struct {
			Name string
			Time int64
		}{nameStr, p.Time / p.Count})
		return true
	})
	// sort by time using sort.Slice
	sort.Slice(ptpair, func(i, j int) bool {
		return ptpair[i].Time > ptpair[j].Time
	})

	// dump ptpair with profiles
	for _, pt := range ptpair {
		pany, _ := profiles.Load(pt.Name)
		p := pany.(*Profile)
		perTime := time.Duration(pt.Time).String()
		totalTime := time.Duration(p.Time).String()
		println(pt.Name, perTime, totalTime, p.Count)
	}
}

func registerProfSignalHandler() {
	// dump on USR1
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR1)
		for {
			<-c
			dumpProfiles()
		}
	}()
	// clear on USR2
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR2)
		for {
			<-c
			profiles = sync.Map{}
		}
	}()
}

// contextに値をセットするmiddlewareの例
func ProfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime(r.URL.String())
		next.ServeHTTP(w, r)
		endTime(r.URL.String())
	})
}
