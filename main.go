package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//
// Based on Docs at: https://github.com/vloschiavo/powerwall2
//

type Record struct {
	LastCommunicationTime time.Time `json:"last_communication_time"`
	InstantPower          float64   `json:"instant_power"`
	InstantReactivePower  float64   `json:"instant_reactive_power"`
	InstantApparentPower  float64   `json:"instant_apparent_power"`
	Frequency             float64   `json:"frequency"`
	EnergyExported        float64   `json:"energy_exported"`
	EnergyImported        float64   `json:"energy_imported"`
	InstantAverageVoltage float64   `json:"instant_average_voltage"`
	InstantTotalCurrent   float64   `json:"instant_total_current"`
	Timeout               int       `json:"timeout"`
}

type PowerwallStatus struct {
	Site    Record `json:"site"` // This is really the "Grid"
	Battery Record `json:"battery"`
	Load    Record `json:"load"`
	Solar   Record `json:"solar"`
}

type StateOfEnergy struct {
	Percentage float64 `json:"percentage"`
}

const (
	Prefix = "tesla_powerwall"
)

var Sources = []string{"site", "battery", "load", "solar"}

func queryStateOfEnergy(host string) (*StateOfEnergy, error) {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Basic HTTP GET request
	resp, err := client.Get(fmt.Sprintf("https://%s/api/system_status/soe", host))
	if err != nil {
		return nil, errors.Wrap(err, "getting http response from Powerwall API")
	}
	defer resp.Body.Close()

	// Read body from response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "readinr http response from Powerwall API")
	}

	status := &StateOfEnergy{}

	if err = json.Unmarshal(body, status); err != nil {
		return nil, errors.Wrap(err, "parsing JSON response from Powerwall API")
	}

	fmt.Printf("%+v\n", status)
	return status, nil
}

func queryMeters(host string) (*PowerwallStatus, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Basic HTTP GET request
	resp, err := client.Get(fmt.Sprintf("https://%s/api/meters/aggregates", host))
	if err != nil {
		return nil, errors.Wrap(err, "getting http response from Powerwall API")
	}
	defer resp.Body.Close()

	// Read body from response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "readinr http response from Powerwall API")
	}

	status := &PowerwallStatus{}

	if err = json.Unmarshal(body, status); err != nil {
		return nil, errors.Wrap(err, "parsing JSON response from Powerwall API")
	}

	fmt.Printf("%+v\n", status)
	return status, nil

}

func populateSource(source string, rec Record, reg *prometheus.Registry) error {

	instantPower := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_instant_power", Prefix),
			Help: "Instant power for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(instantPower); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			instantPower = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_power metric already registered")
		}
	}

	instantPower.WithLabelValues(source).Set(rec.InstantPower)

	instantReactivePower := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_instant_reactive_power", Prefix),
			Help: "Instant reactive power for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(instantReactivePower); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			instantReactivePower = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	instantReactivePower.WithLabelValues(source).Set(rec.InstantReactivePower)

	instantApparentPower := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_instant_apparent_power", Prefix),
			Help: "Instant reactive power for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(instantApparentPower); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			instantApparentPower = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	instantApparentPower.WithLabelValues(source).Set(rec.InstantReactivePower)

	frequency := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_frequency", Prefix),
			Help: "Frequency for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(frequency); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			frequency = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	frequency.WithLabelValues(source).Set(rec.Frequency)

	energyExported := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_energy_exported", Prefix),
			Help: "Frequency for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(energyExported); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			energyExported = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	energyExported.WithLabelValues(source).Set(rec.EnergyExported)

	energyImported := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_energy_imported", Prefix),
			Help: "Frequency for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(energyImported); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			energyImported = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	energyImported.WithLabelValues(source).Set(rec.EnergyImported)

	instantAverageVoltage := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_instant_average_voltage", Prefix),
			Help: "Frequency for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(instantAverageVoltage); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			instantAverageVoltage = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	instantAverageVoltage.WithLabelValues(source).Set(rec.InstantAverageVoltage)

	instantTotalCurrent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_instant_total_current", Prefix),
			Help: "Frequency for source",
		},
		[]string{"source"},
	)

	if err := reg.Register(instantTotalCurrent); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			instantTotalCurrent = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			return errors.Wrap(err, "handling instant_reactive_power metric already registered")
		}
	}

	instantTotalCurrent.WithLabelValues(source).Set(rec.InstantTotalCurrent)

	return nil

}

func generateMetricHandler() func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		target := r.URL.Query().Get("target")
		if target == "" {
			w.Write([]byte("You must provide a target parameter."))
			w.WriteHeader(http.StatusBadRequest)
		}

		status, err := queryMeters(target)
		if err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		reg := prometheus.NewRegistry()

		if err = populateSource("site", status.Site, reg); err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err = populateSource("battery", status.Battery, reg); err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err = populateSource("load", status.Load, reg); err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err = populateSource("solar", status.Solar, reg); err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		soe, err := queryStateOfEnergy(target)
		if err != nil {
			log.Printf("%+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		battery := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_battery_percentage", Prefix),
				Help: "Battery percentage of capacity",
			},
		)

		reg.Register(battery)
		battery.Set(soe.Percentage)

		h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		})
		h.ServeHTTP(w, r)

	}
}

func main() {

	promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	)
	http.HandleFunc("/probe", generateMetricHandler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Listening on 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
