package cmd

import (
	"fmt"
	"infiniband_exporter/ibdiagnet2"
	iblog "infiniband_exporter/log"
	"infiniband_exporter/util"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	LogPath   string
	HttpPort  int
	RunMode   string
	WorkDir   string
	GetConfig bool
	SyncData  = new(ibdiagnet2.SyncSwitchData)
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
			configPath := fmt.Sprintf("%s/config", WorkDir)
			if !GetConfig {
				util.SetCache(filepath.Join(configPath, "config.yaml"))
			}
			if RunMode == "prod" {
				SyncData = GetSyncSwitchDataConfig()
				iblog.GetLogger().Info(fmt.Sprintf("SyncSwitchData: %v", SyncData))
			}
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
		"infiniband_exporter.log", "a string parameter",
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
		"dev",
		"an string parameter[dev prod]",
	)
	rootCmd.Flags().StringVarP(
		&WorkDir,
		"workdir",
		"w",
		"/Users/xlmh/Code/github/infiniband_exporter",
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
	if RunMode == "prod" {
		if _, err := SyncData.SyncSwitchData(); err != nil {
			iblog.GetLogger().Error(fmt.Sprintf("SyncSwitchData error: %s", err))
			return
		} else {
			_, err := util.ExecCmd(
				"tar", "-xzvf", fmt.Sprintf("%s/data/ib.tgz", WorkDir), "-C", fmt.Sprintf("%s/data/ibdiagnet2", WorkDir),
			)
			if err != nil {
				iblog.GetLogger().Error(fmt.Sprintf("tar zxvf  ib.tgz error: %s", err))
				return
			}

		}
	} else {
		iblog.GetLogger().Info("RunMode is dev, no need to sync data")
	}
	linkNetDump := ibdiagnet2.LinkNetDump{
		FilePath: filepath.Join(
			fmt.Sprintf("%s/data/ibdiagnet2", WorkDir),
			"ibdiagnet2.net_dump",
		),
		ConfigPath: filepath.Join(
			fmt.Sprintf("%s/config", WorkDir),
			"config.yaml",
		),
		GetConfig: GetConfig,
	}
	var dumper ibdiagnet2.Dumper = &linkNetDump
	dumper.UpdateMetrics()

	linkPm := ibdiagnet2.LinkPm{
		FilePath: filepath.Join(
			fmt.Sprintf("%s/data/ibdiagnet2", WorkDir),
			"ibdiagnet2.pm",
		),
	}

	var pmer ibdiagnet2.Pmer = &linkPm
	pmer.UpdateMetrics()

	promhttp.Handler().ServeHTTP(w, r)
}

func GetSyncSwitchDataConfig() *ibdiagnet2.SyncSwitchData {
	type Config struct {
		SyncDataConfig ibdiagnet2.SyncSwitchData `yaml:"syncDataConfig"`
	}
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("%s/config", WorkDir))
	if err := viper.ReadInConfig(); err != nil {
		iblog.GetLogger().Error(fmt.Sprintf("Error reading sync data config file, %s", err))
		os.Exit(1)
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		iblog.GetLogger().Error(fmt.Sprintf("Unable to decode into struct, %v", err))
		os.Exit(1)
	}
	return &config.SyncDataConfig
}
