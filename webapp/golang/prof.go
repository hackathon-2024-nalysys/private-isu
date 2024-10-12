package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Profile struct {
	LastStart int64
	Time 	int64
	Count 	int64
}

// ここにためる
var profiles = map[string]*Profile{}

// はかりはじめる
func startTime(name string) {
	p := profiles[name]

	if p == nil {
		p = &Profile{}
		profiles[name] = p
	}
	 
	println("start", name)
	// get current time
	p.LastStart = time.Now().UnixNano()
}

// はかりおわる
func endTime(name string) {
	p := profiles[name]

	if p == nil {
		p = &Profile{}
		profiles[name] = p
	}

	// get current time
	p.Time += time.Now().UnixNano() - p.LastStart
	p.Count++
}

func dumpProfiles() {
	println("---DUMP PROFILE---")
	for name, p := range profiles {
		// format time pretty
		time := time.Duration(p.Time).String()
		println(name, time, p.Count)
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
			profiles = map[string]*Profile{}
		}
	}()
}


// contextに値をセットするmiddlewareの例
func ProfMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // ctx := context.WithValue(r.Context(), "user", "123")
		startTime(r.URL.Path)
    next.ServeHTTP(w, r)
		endTime(r.URL.Path)
  })
}
