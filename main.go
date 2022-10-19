package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Gauge
var onlineUsers = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "goapp_online_users", // nome da métrica
	Help: "Online users",       // decrição da métrica
	ConstLabels: map[string]string{
		"course": "fullcycle", // podemos colocar nome que quiser
	},
})

// Counter
var httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "goapp_http_requests_total",
	Help: "Count of all HTTP requests for goapp",
}, []string{})

// Histogram
var httpDurantion = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "goapp_http_requests_duration",
	Help: "Duration in seconds of all HTTP requests",
}, []string{"handler"})

func main() {
	r := prometheus.NewRegistry()
	r.MustRegister(onlineUsers)
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpDurantion)

	go func() {
		for {
			onlineUsers.Set(float64(rand.Intn(2000)))
		}
	}()

	home := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(4)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello Full Cycle"))
	})

	contact := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello Full Cycle | Contact"))
	})

	d := promhttp.InstrumentHandlerDuration(
		httpDurantion.MustCurryWith(prometheus.Labels{"handler": "home"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, home),
	)

	d2 := promhttp.InstrumentHandlerDuration(
		httpDurantion.MustCurryWith(prometheus.Labels{"handler": "contact"}),
		promhttp.InstrumentHandlerCounter(httpRequestsTotal, contact),
	)

	// acessando o App na rota "/" será incrementado a métrica de contador ao executar o handler home
	// http.Handle("/", promhttp.InstrumentHandlerCounter(httpRequestsTotal, home))

	// agora a home com a duration
	http.Handle("/", d)
	http.Handle("/contact", d2)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8181", nil))
}
