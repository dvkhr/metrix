package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dvkhr/metrix.git/internal/metric"
	"github.com/dvkhr/metrix.git/internal/storage"
)

var ms storage.MemStorage

func gaugeMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	n := req.PathValue("name")
	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}

	/*Debug output*/
	sv, _ := ms.GetGaugeMetric(n)
	fmt.Printf("BEFORE: Metric %s has value %g\n", n, sv)

	v, err := strconv.ParseFloat(req.PathValue("value"), 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	ms.PutGaugeMetric(req.PathValue("name"), metric.GaugeMetricValue(v))

	/*Debug output*/
	sv, _ = ms.GetGaugeMetric(n)
	fmt.Printf("AFTER: Metric %s has value %v\n", n, sv)
	res.Write([]byte("Done!"))
}

func counterMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	n := req.PathValue("name")

	/*Debug output*/
	sv, _ := ms.GetCounterMetric(n)
	fmt.Printf("BEFORE: Metric %s has value %v\n", n, sv)

	if len(n) == 0 {
		http.Error(res, "Incorrect name!", http.StatusNotFound)
		return
	}
	v, err := strconv.ParseInt(req.PathValue("value"), 10, 64)
	if err != nil {
		http.Error(res, "Incorrect value!", http.StatusBadRequest)
		return
	}
	ms.PutCounterMetric(req.PathValue("name"), metric.CounterMetricValue(v))

	/*Debug output*/
	sv, _ = ms.GetCounterMetric(n)
	fmt.Printf("AFTER: Metric %s has value %v\n", n, sv)
	res.Write([]byte("Done!"))
}

func incorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

func notfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
}

func main() {
	ms.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/{name}/{value}", gaugeMetric)
	mux.HandleFunc("/update/counter/{name}/{value}", counterMetric)
	mux.HandleFunc("/update/gauge/", notfoundMetricRq)
	mux.HandleFunc("/update/counter/", notfoundMetricRq)
	mux.HandleFunc("/update/", incorrectMetricRq)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}
