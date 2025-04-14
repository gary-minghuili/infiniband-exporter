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
	labels                    = util.GetFieldNames(NetDump{})
	infinibandLinkInfoCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "infiniband_link_info_total",
			Help: "Total infiniband link info",
		},
		labels,
	)
	infinibandLinkInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_info_state",
			Help: "Gauge infiniband link info",
		},
		labels,
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
	remote_guid string
	remote_name string
	remote_port string
	state       string
	local_guid  string
	local_name  string
	local_port  string
}

func init() {
	prometheus.MustRegister(infinibandLinkInfoCounter)
	prometheus.MustRegister(infinibandLinkInfoGauge)
}

func (d *LinkNetDump) GetContent(filepath string) (*[]string, error) {
	file_content, err := util.ReadFileContent(filepath)
	if err != nil {
		log.GetLogger().Error("read file error")
	}
	re := regexp.MustCompile(`(?m)(.*),\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`)
	indexes := re.FindAllStringIndex(file_content, -1)
	var blocks []string
	for i, match := range indexes {
		if i == len(indexes)-1 {
			blocks = append(blocks, strings.TrimSpace(file_content[match[0]:]))
		} else {
			blocks = append(blocks, strings.TrimSpace(file_content[match[0]:indexes[i+1][0]]))
		}
	}
	return &blocks, nil
}

func (d *LinkNetDump) ParseContent(blocks *[]string) (*[]NetDump, error) {
	var netdumps []NetDump
	for _, block := range *blocks {
		switch_expr := `(?m)"(.*)",\s(\w+),\s(0x\w{16}),\sLID\s(\d+)`
		switch_match, err := regexp.Compile(switch_expr)
		if err != nil {
			log.GetLogger().Error("ReSwitch Error compiling regex")
			return nil, err
		}
		sub_switch_match := switch_match.FindStringSubmatch(block)
		remote_name := sub_switch_match[1]
		remote_guid := sub_switch_match[3]
		// remote_port := subSwitchMatch[4]
		active_expr := `\s+(\d+/\d+/\d+)\s+:\s(\d+)\s+:\s(\w+)\s+:\s+(\w+\s\w+)\s+:\s+(\d+)\s+:\s+(\d+\w+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(.*)\s+:\s+(\w{18})\s+:\s+(\w+/\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+"(.*)\s(\w+)"`
		active_match, err := regexp.Compile(active_expr)
		if err != nil {
			log.GetLogger().Error("sub_switch_match error compiling regex")
			return nil, err
		}
		sub_active_match := active_match.FindAllStringSubmatch(block, -1)
		for _, match := range sub_active_match {
			var local_name string
			if value, exists := global.Hca_mlx_map[match[14]]; exists {
				local_name = fmt.Sprintf(`%s %s`, match[13], value)
			} else {
				local_name = fmt.Sprintf(`%s %s`, match[13], match[14])
			}
			netdump := NetDump{
				remote_guid: remote_guid,
				remote_name: remote_name,
				remote_port: match[2],
				state:       match[3],
				local_guid:  match[10],
				local_name:  local_name,
				local_port:  "",
			}
			netdumps = append(netdumps, netdump)
		}
		down_expr := `(\d+/\d+/\d+)\s+:\s+(\d+)\s+:\s+(\w+)\s+:\s+(\w+).*N/A.*`
		down_match, err := regexp.Compile(down_expr)
		if err != nil {
			log.GetLogger().Error("netdump down error compiling regex")
			return nil, err
		}
		sub_down_match := down_match.FindAllStringSubmatch(block, -1)
		for _, match := range sub_down_match {
			netdump := NetDump{
				remote_guid: remote_guid,
				remote_name: remote_name,
				remote_port: match[2],
				state:       match[3],
				local_guid:  "",
				local_name:  "", // TODO
				local_port:  "",
			}
			netdumps = append(netdumps, netdump)
		}
	}
	return &netdumps, nil
}

func (d *LinkNetDump) UpdateMetrics(netDump *[]NetDump) {
	var value float64
	for _, net := range *netDump {
		infinibandLinkInfoCounter.WithLabelValues(
			net.remote_guid,
			net.remote_name,
			net.remote_port,
			net.state,
			net.local_guid,
			net.local_name,
			net.local_port,
		).Inc()

		if net.state == "ACT" {
			value = 1
		} else {
			value = 0
		}
		infinibandLinkInfoGauge.WithLabelValues(
			net.remote_guid,
			net.remote_name,
			net.remote_port,
			net.state,
			net.local_guid,
			net.local_name,
			net.local_port,
		).Set(value)
	}
}
