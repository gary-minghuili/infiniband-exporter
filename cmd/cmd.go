package cmd

import (
	"fmt"
	iblog "infiniband/log"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	logPath        string
	configFilePath string
	httpPort       int
	runMode        string
)

func NewInfinibandExporterCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "infiniband-exporter",
		Short: "infiniband-exporter -p 9690 -l /var/log/infiniband-exporter.log",
		Long:  `infiniband-exporter -port 9690 -log /var/log/infiniband-exporter.log`,
		RunE: func(cmd *cobra.Command, args []string) error {
			httpPort, _ = cmd.Flags().GetInt("port")
			logPath, _ = cmd.Flags().GetString("log")
			runMode, _ = cmd.Flags().GetString("mode")
			err := iblog.InitLogger(logPath)
			if err != nil {
				log.Fatalf("Failed to initialize logger: %v", err)
			}
			iblog.GetLogger().Info("Starting server......")
			http.Handle("/metrics", http.HandlerFunc(MetricsHandler))
			err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", httpPort), nil)
			if err != nil {
				iblog.GetLogger().Error("http.ListenAndServe error")
				panic(err)
			}
			return nil
		},
	}
	rootCmd.Flags().StringVarP(&logPath, "log", "l", "/var/log/infiniband-exporter.log", "a string parameter")
	rootCmd.Flags().IntVarP(&httpPort, "port", "p", 9690, "an integer parameter")
	rootCmd.Flags().StringVarP(&runMode, "mode", "m", "prod", "an string parameter[dev prod]")
	return rootCmd
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	promhttp.Handler().ServeHTTP(w, r)
}
