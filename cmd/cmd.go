package cmd

import (
	"fmt"
	"infiniband_exporter/ibdiagnet2"
	iblog "infiniband_exporter/log"
	"infiniband_exporter/util"
	"log"
	"net/http"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	LogPath   string
	HttpPort  int
	RunMode   string
	WorkDir   string
	GetConfig bool
)

func NewInfinibandExporterCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "infiniband-exporter",
		Short: "infiniband-exporter -p 9690 -l /var/log/infiniband-exporter.log",
		Long:  `infiniband-exporter -port 9690 -log /var/log/infiniband-exporter.log`,
		RunE: func(cmd *cobra.Command, args []string) error {
			HttpPort, _ = cmd.Flags().GetInt("port")
			LogPath, _ = cmd.Flags().GetString("log")
			RunMode, _ = cmd.Flags().GetString("mode")
			WorkDir, _ = cmd.Flags().GetString("workdir")
			GetConfig, _ = cmd.Flags().GetBool("getconfig")
			err := iblog.InitLogger(LogPath)
			if err != nil {
				log.Fatalf("Failed to initialize logger: %v", err)
			}
			iblog.GetLogger().Info("Starting server......")
			util.SetCache(filepath.Join(
				fmt.Sprintf("%sconfig", WorkDir),
				"config.yaml",
			))
			http.Handle("/metrics", http.HandlerFunc(MetricsHandler))
			err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", HttpPort), nil)
			if err != nil {
				iblog.GetLogger().Error("http.ListenAndServe error")
				panic(err)
			}
			return nil
		},
	}
	rootCmd.Flags().StringVarP(
		&LogPath,
		"log",
		"l",
		"/var/log/infiniband-exporter.log", "a string parameter",
	)
	rootCmd.Flags().IntVarP(
		&HttpPort,
		"port",
		"p",
		9690,
		"an integer parameter",
	)
	rootCmd.Flags().StringVarP(
		&RunMode,
		"mode",
		"m",
		"prod",
		"an string parameter[dev prod]",
	)
	rootCmd.Flags().StringVarP(
		&WorkDir,
		"workdir",
		"w",
		"/Users/xlmh/Code/github/infiniband_exporter/",
		"an string parameter",
	)
	rootCmd.Flags().BoolVarP(
		&GetConfig,
		"getconfig",
		"g",
		false,
		"an bool parameter",
	)
	return rootCmd
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	linkNetDump := ibdiagnet2.LinkNetDump{
		FilePath: filepath.Join(
			fmt.Sprintf("%sdata/ibdiagnet2", WorkDir),
			"ibdiagnet2.net_dump",
		),
		ConfigPath: filepath.Join(
			fmt.Sprintf("%sconfig", WorkDir),
			"config.yaml",
		),
		GetConfig: GetConfig,
	}
	var dumper ibdiagnet2.Dumper = &linkNetDump
	dumper.UpdateMetrics()

	linkPm := ibdiagnet2.LinkPm{
		FilePath: filepath.Join(
			fmt.Sprintf("%sdata/ibdiagnet2", WorkDir),
			"ibdiagnet2.pm",
		),
	}

	var pmer ibdiagnet2.Pmer = &linkPm
	pmer.UpdateMetrics()

	promhttp.Handler().ServeHTTP(w, r)
}
