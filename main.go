package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/ibmdb/go_ibm_db"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	// Version will be set at build time.
	Version            = "0.0.1.dev"
	listenAddress      = flag.String("web.listen-address", ":9161", "Address to listen on for web interface and telemetry.")
	metricPath         = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	landingPage        = []byte("<html><head><title>IBM DB2 Exporter " + Version + "</title></head><body><h1>IBM DB2 Exporter " + Version + "</h1><p><a href='" + *metricPath + "'>Metrics</a></p></body></html>")
	defaultFileMetrics = flag.String("default.metrics", "default-metrics.toml", "File with default metrics in a TOML file.")
	customMetrics      = flag.String("custom.metrics", os.Getenv("CUSTOM_METRICS"), "File that may contain various custom metrics in a TOML file.")
	queryTimeout       = flag.String("query.timeout", "5", "Query timeout (in seconds).")
	db2dsn             = flag.String("dsn", os.Getenv("DB_DSN"), "Default DSN:DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;")
)

// Metric name parts.
const (
	namespace = "ibmdb2"
	exporter  = "exporter"
)

// Metric object description
type Metric struct {
	Context          string
	Labels           []string
	MetricsDesc      map[string]string
	MetricsType      map[string]string
	FieldToAppend    string
	Request          string
	IgnoreZeroResult bool
}

// Metrics Used to load multiple metrics from file
type Metrics struct {
	Metric []Metric
}

// Metrics to scrap. Use external file (default-metrics.toml and custom if provided)
var (
	metricsToScrap    Metrics
	additionalMetrics Metrics
)

// Exporter collects IBM DB2 metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn             string
	duration, error prometheus.Gauge
	totalScrapes    prometheus.Counter
	scrapeErrors    *prometheus.CounterVec
	up              prometheus.Gauge
	db              *sql.DB
}

// NewExporter returns a new IBM DB2 exporter for the provided DSN.
func NewExporter(dsn string) *Exporter {
	db, err := sql.Open("go_ibm_db", dsn)
	if err != nil {
		log.Errorln("Error while connecting to", dsn)
		panic(err)
	}
	return &Exporter{
		dsn: dsn,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from IBM DB2.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrapes_total",
			Help:      "Total number of times IBM DB2 was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occured scraping a Oracle database.",
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: exporter,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from IBM DB2 resulted in an error (1 for error, 0 for success).",
		}),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the IBM DB2 database server is up.",
		}),
		db: db,
	}
}

// Describe describes all the metrics exported by the exporter.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh

}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)
	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
	ch <- e.up
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	var err error
	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
		if err == nil {
			e.error.Set(0)
		} else {
			e.error.Set(1)
		}
	}(time.Now())

	// Noop function for simple SELECT 1 FROM SYSIBM.SYSDUMMY1
	noop := func(row map[string]string) error { return nil }
	if err = GeneratePrometheusMetrics(e.db, noop, "SELECT 1 FROM SYSIBM.SYSDUMMY1"); err != nil {
		log.Errorln("Error pinging DB2:", err)
		// close old connection
		e.db.Close()
		// Maybe DB2 instance was restarted => try to reconnect
		log.Infoln("Try to reconnect...")
		e.db, err = sql.Open("go_ibm_db", e.dsn)
		if err != nil {
			log.Errorln("Error while connecting to DB2:", err)
			e.up.Set(0)
			return
		}
		if err = GeneratePrometheusMetrics(e.db, noop, "SELECT 1 FROM SYSIBM.SYSDUMMY1"); err != nil {
			log.Error("Unable to connect to DB2:", err)
			e.up.Set(0)
			return
		}
	}
	e.up.Set(1)

	for _, metric := range metricsToScrap.Metric {
		if err = ScrapeMetric(e.db, ch, metric); err != nil {
			log.Errorln("Error scraping for", metric.Context, ":", err)
			e.scrapeErrors.WithLabelValues(metric.Context).Inc()
		}
	}

}

// GetMetricType get type
func GetMetricType(metricType string, metricsType map[string]string) prometheus.ValueType {
	var strToPromType = map[string]prometheus.ValueType{
		"gauge":   prometheus.GaugeValue,
		"counter": prometheus.CounterValue,
	}

	strType, ok := metricsType[strings.ToLower(metricType)]
	if !ok {
		return prometheus.GaugeValue
	}
	valueType, ok := strToPromType[strings.ToLower(strType)]
	if !ok {
		panic(errors.New("Error while getting prometheus type " + strings.ToLower(strType)))
	}
	return valueType
}

// ScrapeMetric interface method to call ScrapeGenericValues using Metric struct values
func ScrapeMetric(db *sql.DB, ch chan<- prometheus.Metric, metricDefinition Metric) error {
	return ScrapeGenericValues(db, ch, metricDefinition.Context, metricDefinition.Labels,
		metricDefinition.MetricsDesc, metricDefinition.MetricsType,
		metricDefinition.FieldToAppend, metricDefinition.IgnoreZeroResult,
		metricDefinition.Request)
}

// ScrapeGenericValues generic method for retrieving metrics.
func ScrapeGenericValues(db *sql.DB, ch chan<- prometheus.Metric, context string, labels []string,
	metricsDesc map[string]string, metricsType map[string]string, fieldToAppend string, ignoreZeroResult bool, request string) error {
	metricsCount := 0
	genericParser := func(row map[string]string) error {
		// Construct labels value
		labelsValues := []string{}
		for _, label := range labels {
			labelsValues = append(labelsValues, row[label])
		}
		//debug

		//fmt.Println("label-debug", labelsValues)
		log.Debugln("label-debug", labelsValues)
		// Construct Prometheus values to sent back
		for metric, metricHelp := range metricsDesc {
			metric = strings.ToLower(metric)
			value, err := strconv.ParseFloat(strings.TrimSpace(row[metric]), 64)

			//debug
			//fmt.Println("metric-debug", metric, ":", value)
			log.Debugln("metric-debug", metric, ":", value)

			// If not a float, skip current metric
			if err != nil {
				continue
			}
			// If metric do not use a field content in metric's name
			if strings.Compare(fieldToAppend, "") == 0 {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, metric),
					metricHelp,
					labels, nil,
				)
				ch <- prometheus.MustNewConstMetric(desc, GetMetricType(metric, metricsType), value, labelsValues...)
				// If no labels, use metric name
			} else {
				desc := prometheus.NewDesc(
					prometheus.BuildFQName(namespace, context, cleanName(row[fieldToAppend])),
					metricHelp,
					nil, nil,
				)
				ch <- prometheus.MustNewConstMetric(desc, GetMetricType(metric, metricsType), value)
			}
			metricsCount++
		}
		return nil
	}
	err := GeneratePrometheusMetrics(db, genericParser, request)
	if err != nil {
		return err
	}

	if !ignoreZeroResult && metricsCount == 0 {
		return errors.New("No metrics found while parsing")
	}
	return err
}

// GeneratePrometheusMetrics Parse SQL result and call parsing function to each row
func GeneratePrometheusMetrics(db *sql.DB, parse func(row map[string]string) error, query string) error {

	// Add a timeout
	timeout, err := strconv.Atoi(*queryTimeout)
	if err != nil {
		log.Fatal("error while converting timeout option value: ", err)
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)

	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("DB2 query timed out")
	}

	if err != nil {
		return err
	}
	cols, err := rows.Columns()
	defer rows.Close()

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		//columns := make([]interface{}, len(cols))
		columns := make([]sql.RawBytes, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]string)
		for i, val := range columns {
			if m[strings.ToLower(cols[i])] = string(val); val == nil {
				m[strings.ToLower(cols[i])] = "NULL"
			}
		}

		// fmt.Println(m)
		log.Debugln(m)
		// Call function to parse row
		if err := parse(m); err != nil {
			return err
		}
	}

	return nil

}

// DB2 gives us some ugly names back. This function cleans things up for Prometheus.
func cleanName(s string) string {
	s = strings.Replace(s, " ", "_", -1) // Remove spaces
	s = strings.Replace(s, "(", "", -1)  // Remove open parenthesis
	s = strings.Replace(s, ")", "", -1)  // Remove close parenthesis
	s = strings.Replace(s, "/", "", -1)  // Remove forward slashes
	s = strings.ToLower(s)
	return s
}

func main() {
	flag.Parse()
	log.Infoln("Starting ibmdb2_exporter " + Version)
	dsn := *db2dsn
	if dsn == "" {
		dsn = "DATABASE=sample; HOSTNAME=localhost; PORT=60000; PROTOCOL=TCPIP; UID=db2inst1; PWD=db2inst1;"
		log.Infoln("With default DSN config. To change it set ENV DB2_DSN or -dsn flag.")
	}
	log.Infoln("Running with DB2_DSN=", dsn)
	// Load default metrics
	if _, err := toml.DecodeFile(*defaultFileMetrics, &metricsToScrap); err != nil {
		log.Errorln(err)
		panic(errors.New("Error while loading " + *defaultFileMetrics))
	}

	// If custom metrics, load it
	if strings.Compare(*customMetrics, "") != 0 {
		if _, err := toml.DecodeFile(*customMetrics, &additionalMetrics); err != nil {
			log.Errorln(err)
			panic(errors.New("Error while loading " + *customMetrics))
		}
		metricsToScrap.Metric = append(metricsToScrap.Metric, additionalMetrics.Metric...)
	}
	exporter := NewExporter(dsn)
	prometheus.MustRegister(exporter)
	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
