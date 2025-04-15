package cmd

import (
	"fmt"
	"infiniband_exporter/ibdiagnet2"
	iblog "infiniband_exporter/log"
	"log"
	"net/http"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	LogPath        string
	ConfigFilePath string
	HttpPort       int
	RunMode        string
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
			err := iblog.InitLogger(LogPath)
			if err != nil {
				log.Fatalf("Failed to initialize logger: %v", err)
			}
			iblog.GetLogger().Info("Starting server......")
			http.Handle("/metrics", http.HandlerFunc(MetricsHandler))
			err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", HttpPort), nil)
			if err != nil {
				iblog.GetLogger().Error("http.ListenAndServe error")
				panic(err)
			}
			return nil
		},
	}
	rootCmd.Flags().StringVarP(&LogPath, "log", "l", "/var/log/infiniband-exporter.log", "a string parameter")
	rootCmd.Flags().IntVarP(&HttpPort, "port", "p", 9690, "an integer parameter")
	rootCmd.Flags().StringVarP(&RunMode, "mode", "m", "prod", "an string parameter[dev prod]")
	return rootCmd
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	linkNetDump := ibdiagnet2.LinkNetDump{
		FilePath: filepath.Join(
			"/Users/xlmh/Code/github/infiniband_exporter/data/ibdiagnet2",
			"ibdiagnet2.net_dump",
		),
	}
	var dumper ibdiagnet2.Dumper = &linkNetDump
	file_content, err := dumper.GetContent(linkNetDump.FilePath)
	if err != nil {
		panic(err)
	}
	netDumps, err := dumper.ParseContent(file_content)
	if err != nil {
		panic(err)
	}
	dumper.UpdateMetrics(netDumps)

	linkPm := ibdiagnet2.LinkPm{
		FilePath: filepath.Join(
			"/Users/xlmh/Code/github/infiniband_exporter/data/ibdiagnet2",
			"ibdiagnet2.pm",
		),
	}

	var pmer ibdiagnet2.Pmer = &linkPm
	fileContent, err := pmer.GetContent(linkPm.FilePath)
	if err != nil {
		panic(err)
	}
	pmer.UpdateMetrics(fileContent)

	promhttp.Handler().ServeHTTP(w, r)
}
