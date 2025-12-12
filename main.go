package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ApplicationName = "prometheus_varnish_exporter"
	Version         string
	VersionHash     string
	VersionDate     string

	PrometheusExporter = NewPrometheusExporter()
	VarnishVersion     = NewVarnishVersion()
	ExitHandler        = &exitHandler{}

	StartParams = &startParams{
		ListenAddress:  ":9131", // Reserved and publicly announced at https://github.com/prometheus/prometheus/wiki/Default-port-allocations
		Path:           "/metrics",
		VarnishstatExe: "varnishstat",
		Params:         &varnishstatParams{},
		LogLevel:       "info",
	}
)

type startParams struct {
	ListenAddress          string
	Path                   string
	HealthPath             string
	VarnishstatExe         string
	VarnishDockerContainer string
	Params                 *varnishstatParams

	Verbose       bool
	ExitOnErrors  bool
	Test          bool
	Raw           bool
	WithGoMetrics bool

	// logging
	LogLevel string
	LogJSON  bool

	noExit bool // deprecated
}

type varnishstatParams struct {
	Instance string
	VSM      string
}

func (p *varnishstatParams) isEmpty() bool {
	return p.Instance == "" && p.VSM == ""
}

func (p *varnishstatParams) make() (params []string) {
	// -n
	if p.Instance != "" {
		params = append(params, "-n", p.Instance)
	}
	// -N is not supported by 3.x
	if p.VSM != "" && VarnishVersion.EqualsOrGreater(4, 0) {
		params = append(params, "-N", p.VSM)
	}
	return params
}

func main() {
	// prometheus conventions
	flag.StringVar(&StartParams.ListenAddress, "web.listen-address", StartParams.ListenAddress, "Address on which to expose metrics and web interface.")
	flag.StringVar(&StartParams.Path, "web.telemetry-path", StartParams.Path, "Path under which to expose metrics.")
	flag.StringVar(&StartParams.HealthPath, "web.health-path", StartParams.HealthPath, "Path under which to expose healthcheck. Disabled unless configured.")

	// varnish
	flag.StringVar(&StartParams.VarnishstatExe, "varnishstat-path", StartParams.VarnishstatExe, "Path to varnishstat.")
	flag.StringVar(&StartParams.Params.Instance, "n", StartParams.Params.Instance, "varnishstat -n value.")
	flag.StringVar(&StartParams.Params.VSM, "N", StartParams.Params.VSM, "varnishstat -N value.")

	// docker
	flag.StringVar(&StartParams.VarnishDockerContainer, "docker-container-name", StartParams.VarnishDockerContainer, "Docker container name to exec varnishstat in.")

	// logging
	flag.StringVar(&StartParams.LogLevel, "log-level", StartParams.LogLevel, "Log level: debug, info, warn, error")
	flag.BoolVar(&StartParams.LogJSON, "log-json", StartParams.LogJSON, "Log in JSON format")

	// modes
	version := false
	flag.BoolVar(&version, "version", version, "Print version and exit")
	flag.BoolVar(&StartParams.ExitOnErrors, "exit-on-errors", StartParams.ExitOnErrors, "Exit process on scrape errors.")
	flag.BoolVar(&StartParams.Test, "test", StartParams.Test, "Test varnishstat availability, prints available metrics and exits.")
	flag.BoolVar(&StartParams.WithGoMetrics, "with-go-metrics", StartParams.WithGoMetrics, "Export go runtime and http handler metrics")

	// deprecated
	flag.BoolVar(&StartParams.noExit, "no-exit", StartParams.noExit, "Deprecated: see -exit-on-errors")

	flag.Parse()

	if version {
		fmt.Printf("%s %s\n", ApplicationName, getVersion(true))
		os.Exit(0)
	}

	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	if len(StartParams.Path) == 0 || StartParams.Path[0] != '/' {
		logFatal("-web.telemetry-path cannot be empty and must start with a slash '/'", "path", StartParams.Path)
	}
	if len(StartParams.HealthPath) != 0 && StartParams.HealthPath[0] != '/' {
		logFatal("-web.health-path must start with a slash '/' if configured", "path", StartParams.HealthPath)
	}
	if StartParams.Path == StartParams.HealthPath {
		logFatal("-web.telemetry-path and -web.health-path cannot have same value")
	}

	// Don't log warning on !noExit as that would spam for the formed default value.
	if StartParams.noExit {
		slog.Warn("-no-exit is deprecated. As of v1.5 it is the default behavior not to exit process on scrape errors. You can remove this parameter.")
	}

	// Test run or user explicitly wants to exit on any scrape errors during runtime.
	ExitHandler.exitOnError = StartParams.Test == true || StartParams.ExitOnErrors == true

	slog.Info("Initializing application", "applicationName", ApplicationName, "applicationVersion", getVersion(false), "parameters", StartParams)

	// Initialize
	if err := VarnishVersion.Initialize(); err != nil {
		ExitHandler.Errorf("Varnish version initialize failed: %s", err.Error())
	}
	if VarnishVersion.Valid() {
		slog.Info("Found varnishstat.", "varnishVersion", VarnishVersion)
		if err := PrometheusExporter.Initialize(); err != nil {
			logFatal("Prometheus exporter initialize failed", "error", err.Error())
		}
	}

	// Test to verify everything is ok before starting the server
	{
		done := make(chan bool)
		metrics := make(chan prometheus.Metric)
		go func() {
			for m := range metrics {
				if StartParams.Test {
					slog.Info(m.Desc().String())
				}
			}
			done <- true
		}()
		tStart := time.Now()
		buf, err := ScrapeVarnish(metrics)
		close(metrics)
		<-done

		if err == nil {
			slog.Info("Test scrape done", "duration", time.Now().Sub(tStart))
		} else {
			if len(buf) > 0 {
				slog.Debug("Scrape output", "output", string(buf))
			}
			ExitHandler.Errorf("Startup test: %s", err.Error())
		}
	}
	if StartParams.Test {
		return
	}

	// Start serving
	slog.Info("Server starting", "listenAddress", StartParams.ListenAddress, "metricsPath", StartParams.Path)

	if !StartParams.WithGoMetrics {
		registry := prometheus.NewRegistry()
		if err := registry.Register(PrometheusExporter); err != nil {
			logFatal("registry.Register failed: %s", err.Error())
		}
		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorLog: &SlogErrorLogger{slog.Default()},
		})
		http.Handle(StartParams.Path, handler)
	} else {
		prometheus.MustRegister(PrometheusExporter)
		http.Handle(StartParams.Path, promhttp.Handler())
	}

	if StartParams.Path != "/" {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
    <head><title>Varnish Exporter</title></head>
    <body>
        <h1>Varnish Exporter</h1>
    	<p><a href="` + StartParams.Path + `">Metrics</a></p>
    </body>
</html>`))
		})
	}
	if StartParams.HealthPath != "" {
		http.HandleFunc(StartParams.HealthPath, func(w http.ResponseWriter, r *http.Request) {
			// As noted in the "up" metric, needs some way to determine if everything is actually Ok.
			// For now, this just lets us check that we're accepting connections
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Ok")
		})
	}
	logFatalError(http.ListenAndServe(StartParams.ListenAddress, nil))
}

type exitHandler struct {
	sync.RWMutex
	exitOnError bool
	err         error
}

func (ex *exitHandler) Errorf(format string, a ...interface{}) error {
	return ex.Set(fmt.Errorf(format, a...))
}

func (ex *exitHandler) HasError() bool {
	ex.RLock()
	hasError := ex.err != nil
	ex.RUnlock()
	return hasError
}

func (ex *exitHandler) Set(err error) error {
	ex.Lock()
	defer ex.Unlock()

	if err == nil {
		ex.err = nil
		return nil
	}

	errDiffers := ex.err == nil || ex.err.Error() != err.Error()
	ex.err = err

	if ex.exitOnError {
		logFatal("%s", err.Error())
	} else if errDiffers {
		slog.Error(err.Error())
	}
	return err
}

func getVersion(date bool) (version string) {
	if Version == "" {
		return "dev"
	}
	version = fmt.Sprintf("v%s (%s)", Version, VersionHash)
	if date {
		version += " " + VersionDate
	}
	return version
}
