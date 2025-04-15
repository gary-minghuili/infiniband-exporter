package ibdiagnet2

import (
	"fmt"
	"infiniband_exporter/global"
	"infiniband_exporter/log"
	"infiniband_exporter/util"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
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
	GetContent(filepath string) (*[]string, error)
	ParseContent(blocks *[]string) (*[]NetDump, error)
	UpdateMetrics(netDump *[]NetDump)
}

type LinkNetDump struct {
	FilePath string
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

func (d *LinkNetDump) GetContent(filepath string) (*[]string, error) {
	fileContent, err := util.ReadFileContent(filepath)
	if err != nil {
		log.GetLogger().Error("read file error")
	}
	re := regexp.MustCompile(`(?m)(.*),\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`)
	indexes := re.FindAllStringIndex(fileContent, -1)
	var blocks []string
	for i, match := range indexes {
		if i == len(indexes)-1 {
			blocks = append(blocks, strings.TrimSpace(fileContent[match[0]:]))
		} else {
			blocks = append(blocks, strings.TrimSpace(fileContent[match[0]:indexes[i+1][0]]))
		}
	}
	return &blocks, nil
}

func (d *LinkNetDump) ParseContent(blocks *[]string) (*[]NetDump, error) {
	var netdumps []NetDump
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
			netdump := NetDump{
				remoteGuid: remoteGuid,
				remoteName: remoteName,
				remotePort: match[2],
				state:      match[3],
				localGuid:  "",
				localName:  "", // TODO
				localPort:  "",
			}
			netdumps = append(netdumps, netdump)
		}
	}
	return &netdumps, nil
}

func (d *LinkNetDump) UpdateMetrics(netDump *[]NetDump) {
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
