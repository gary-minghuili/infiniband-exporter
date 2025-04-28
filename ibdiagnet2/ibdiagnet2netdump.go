package ibdiagnet2

import (
	"fmt"
	"infiniband_exporter/global"
	"infiniband_exporter/log"
	"infiniband_exporter/util"
	"os"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"
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
	netDumpSwitchInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_switch_info_state",
			Help: "Gauge infiniband switch info",
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
	prometheus.MustRegister(netDumpSwitchInfoGauge)
}

func (d *LinkNetDump) ParseContent() (*[]NetDump, error) {
	var netDumps []NetDump
	configData := make(map[string]any)
	blocks, err := util.GetContent(d.FilePath, `(?m)(.*),\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`)
	if err != nil {
		log.GetLogger().Error("Get content error")
		return nil, err
	}
	remoteNameMap := make(map[string]string)
	for _, block := range *blocks {
		switchExpr := `(?m)"(.*)",\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`
		switchMatch, err := regexp.Compile(switchExpr)
		if err != nil {
			log.GetLogger().Error("Switch error compiling regex")
			return nil, err
		}
		subSwitchMatch := switchMatch.FindStringSubmatch(block)
		remoteName := subSwitchMatch[1]
		remoteGuid := subSwitchMatch[3]
		// remotePort := subSwitchMatch[4]
		activeExprs := []string{
			`\s+(\d+/\d+/\d+)\s+:\s(\d+)\s+:\s(\w+)\s+:\s+(\w+\s\w+)\s+:\s+(\d+)\s+:\s+(\d+\w+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(.*)\s+:\s+(\w{18})\s+:\s+(\w+/\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+"([\w-]+)\s([\w-]+)"`,
			`\s+(\d+/\d+/\d+)\s+:\s(\d+)\s+:\s(\w+)\s+:\s+(\w+\s\w+)\s+:\s+(\d+)\s+:\s+(\d+\w+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(.*)\s+:\s+(\w{18})\s+:\s+(\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+"(.*)"`,
		}
		for _, activeExpr := range activeExprs {
			activeMatch, err := regexp.Compile(activeExpr)
			if err != nil {
				log.GetLogger().Error("Sub switch match error compiling regex")
				return nil, err
			}
			subActiveMatch := activeMatch.FindAllStringSubmatch(block, -1)
			for _, match := range subActiveMatch {
				var localName string
				if len(match) == 15 {
					leafKey := match[14]
					if value, exists := global.HcaMlxMap[match[14]]; exists {
						leafKey = value
						localName = fmt.Sprintf(`%s %s`, match[13], value)
					} else {
						localName = fmt.Sprintf(`%s %s`, match[13], match[14])
					}
					if remoteLeafName, exists := global.MlxLeafMap[leafKey]; exists {
						if _, exists := remoteNameMap[remoteGuid]; !exists {
							remoteNameMap[remoteGuid] = remoteLeafName
						}
					}
				} else {
					localName = match[13]
				}
				netDump := NetDump{
					remoteGuid: remoteGuid,
					remoteName: remoteNameMap[remoteGuid],
					remotePort: match[2],
					state:      match[3],
					localGuid:  match[10],
					localName:  localName,
					localPort:  "",
				}
				netDumps = append(netDumps, netDump)
			}
		}
		downExpr := `(\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(\w+).*N/A.*`
		downMatch, err := regexp.Compile(downExpr)
		if err != nil {
			log.GetLogger().Error("Net dump down error compiling regex")
			return nil, err
		}
		subDownMatch := downMatch.FindAllStringSubmatch(block, -1)
		for _, match := range subDownMatch {
			var localGuid, localName, localPort string
			remotePort, state := match[2], match[3]
			if !d.GetConfig {
				linkMap, exists := util.GetValueFromCache(fmt.Sprintf("%s_%s", remoteGuid, remotePort))
				if exists {
					remoteNameValue, exists := linkMap["remoteName"]
					if exists {
						remoteName = remoteNameValue
					}
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
			}
			netDump := NetDump{
				remoteGuid: remoteGuid,
				remoteName: remoteName,
				remotePort: remotePort,
				state:      state,
				localGuid:  localGuid,
				localName:  localName,
				localPort:  localPort,
			}
			netDumps = append(netDumps, netDump)
		}
	}
	if d.GetConfig {
		for _, netDump := range netDumps {
			configDataKey := fmt.Sprintf("%s_%s", netDump.remoteGuid, netDump.remotePort)
			configData[configDataKey] = map[string]any{
				"remoteName": remoteNameMap[netDump.remoteGuid],
				"remoteGuid": netDump.remoteGuid,
				"remotePort": netDump.remotePort,
				"state":      netDump.state,
				"localGuid":  netDump.localGuid,
				"localName":  netDump.localName,
				"localPort":  netDump.localPort,
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
	return &netDumps, nil
}

func (d *LinkNetDump) UpdateMetrics() {
	netDump, err := d.ParseContent()
	if err != nil {
		log.GetLogger().Error("Parse content error")
		return
	}
	var value float64
	netDumpSwitchs := make(map[string]string, 0)
	for _, net := range *netDump {
		if _, exists := netDumpSwitchs[net.remoteGuid]; !exists {
			netDumpSwitchs[net.remoteGuid] = net.remoteName
		}
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
	netDumpSwitchsFromCache, _ := util.GetKeysFromCache("")
	diffSwitchs := util.DifferenceSlice(netDumpSwitchsFromCache, maps.Keys(netDumpSwitchs))
	for _, remoteGuid := range diffSwitchs {
		if linkMap, exists := util.GetValueFromCache(fmt.Sprintf("%s_1", remoteGuid)); exists {
			netDumpSwitchInfoGauge.WithLabelValues(
				remoteGuid, linkMap["remoteName"], "-",
				"DOWN",
				"-", "-", "-").
				Set(0)
		}
	}
	for remoteGuid, remoteName := range netDumpSwitchs {
		netDumpSwitchInfoGauge.WithLabelValues(
			remoteGuid, remoteName, "-", "UP", "-", "-", "-").
			Set(1)
	}
}
