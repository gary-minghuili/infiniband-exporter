package ibdiagnet2

import (
	"fmt"
	"infiniband_exporter/global"
	"infiniband_exporter/log"
	"infiniband_exporter/util"
	"os"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

var (
	netDumpLabels          = util.GetFieldNames(NetDump{})
	netDumpLinkInfoCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infiniband_link_info_total",
			Help: "Total infiniband link info",
		},
		netDumpLabels,
	)
	netDumpLinkInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_info_state",
			Help: "Gauge infiniband link info",
		},
		netDumpLabels,
	)
)

type Dumper interface {
	ParseContent() (*[]NetDump, error)
	UpdateMetrics()
}

type LinkNetDump struct {
	FilePath   string
	ConfigPath string
	GetConfig  bool
}

type NetDump struct {
	remoteGuid string
	remoteName string
	remotePort string
	state      string
	localGuid  string
	localName  string
	localPort  string
}

func init() {
	prometheus.MustRegister(netDumpLinkInfoCounter)
	prometheus.MustRegister(netDumpLinkInfoGauge)
}

func (d *LinkNetDump) ParseContent() (*[]NetDump, error) {
	var netdumps []NetDump
	configData := make(map[string]any, 0)
	blocks, err := util.GetContent(d.FilePath, `(?m)(.*),\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`)
	if err != nil {
		log.GetLogger().Error("GetContent error")
		return nil, err
	}
	for _, block := range *blocks {
		switchExpr := `(?m)"(.*)",\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`
		switchMatch, err := regexp.Compile(switchExpr)
		if err != nil {
			log.GetLogger().Error("ReSwitch Error compiling regex")
			return nil, err
		}
		subSwitchMatch := switchMatch.FindStringSubmatch(block)
		remoteName := subSwitchMatch[1]
		remoteGuid := subSwitchMatch[3]
		// remotePort := subSwitchMatch[4]
		activeExpr := `\s+(\d+/\d+/\d+)\s+:\s(\d+)\s+:\s(\w+)\s+:\s+(\w+\s\w+)\s+:\s+(\d+)\s+:\s+(\d+\w+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(.*)\s+:\s+(\w{18})\s+:\s+(\w+/\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+"(.*)\s(\w+)"`
		activeMatch, err := regexp.Compile(activeExpr)
		if err != nil {
			log.GetLogger().Error("sub_switch_match error compiling regex")
			return nil, err
		}
		subActiveMatch := activeMatch.FindAllStringSubmatch(block, -1)
		for _, match := range subActiveMatch {
			var localName string
			if value, exists := global.HcaMlxMap[match[14]]; exists {
				localName = fmt.Sprintf(`%s %s`, match[13], value)
			} else {
				localName = fmt.Sprintf(`%s %s`, match[13], match[14])
			}
			netdump := NetDump{
				remoteGuid: remoteGuid,
				remoteName: remoteName,
				remotePort: match[2],
				state:      match[3],
				localGuid:  match[10],
				localName:  localName,
				localPort:  "",
			}
			netdumps = append(netdumps, netdump)
		}
		downExpr := `(\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(\w+).*N/A.*`
		downMatch, err := regexp.Compile(downExpr)
		if err != nil {
			log.GetLogger().Error("netdump down error compiling regex")
			return nil, err
		}
		subDownMatch := downMatch.FindAllStringSubmatch(block, -1)
		for _, match := range subDownMatch {
			var localGuid, localName, localPort string
			remotePort, state := match[2], match[3]
			linkMap, exists := util.GetValueFromCache(fmt.Sprintf("%s_%s", remoteGuid, remotePort))
			if exists {
				localGuidValue, exists := linkMap["localGuid"]
				if exists {
					localGuid = localGuidValue
				}
				localNameValue, exists := linkMap["localName"]
				if exists {
					localName = localNameValue
				}
				localPortValue, exists := linkMap["localPort"]
				if exists {
					localPort = localPortValue
				}
			}
			netdump := NetDump{
				remoteGuid: remoteGuid,
				remoteName: remoteName,
				remotePort: remotePort,
				state:      state,
				localGuid:  localGuid,
				localName:  localName,
				localPort:  localPort,
			}
			netdumps = append(netdumps, netdump)
		}
	}
	if d.GetConfig == true {
		for _, netdump := range netdumps {
			configDataKey := fmt.Sprintf("%s_%s", netdump.remoteGuid, netdump.remotePort)
			configData[configDataKey] = map[string]any{
				"remoteName": netdump.remoteName,
				"remoteGuid": netdump.remoteGuid,
				"remotePort": netdump.remotePort,
				"state":      netdump.state,
				"localGuid":  netdump.localGuid,
				"localName":  netdump.localName,
				"localPort":  netdump.localPort,
			}
		}
		yamlData, err := yaml.Marshal(&configData)
		if err != nil {
			log.GetLogger().Error("Yaml marshal error")
		}
		err = os.WriteFile(d.ConfigPath, yamlData, 0644)
		if err != nil {
			log.GetLogger().Error("Failed to write data into file")
		}
	}
	return &netdumps, nil
}

func (d *LinkNetDump) UpdateMetrics() {
	netDump, err := d.ParseContent()
	if err != nil {
		log.GetLogger().Error("ParseContent error")
		return
	}
	var value float64
	for _, net := range *netDump {
		netDumpLinkInfoCounter.WithLabelValues(
			net.remoteGuid,
			net.remoteName,
			net.remotePort,
			net.state,
			net.localGuid,
			net.localName,
			net.localPort,
		).Inc()

		if net.state == "ACT" {
			value = 1
		} else {
			value = 0
		}
		netDumpLinkInfoGauge.WithLabelValues(
			net.remoteGuid,
			net.remoteName,
			net.remotePort,
			net.state,
			net.localGuid,
			net.localName,
			net.localPort,
		).Set(value)
	}
}
