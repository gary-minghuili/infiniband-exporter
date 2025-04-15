package ibdiagnet2

import (
	"fmt"
	"infiniband_exporter/global"
	"infiniband_exporter/log"
	"infiniband_exporter/util"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	pmLabels      = util.GetFieldNames(Pm{})
	linkDownGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_down_counter",
			Help: "link_down_counter",
		},
		pmLabels,
	)

	linkErrorRecoveryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_error_recovery_counter",
			Help: "link_error_recovery_counter",
		},
		pmLabels,
	)
	symbolErrorCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_symbol_error_counter",
			Help: "symbol_error_counter",
		},
		pmLabels,
	)
	portRcvRemotePhysicalErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_remote_physical_errors",
			Help: "port_rcv_remote_physical_errors",
		},
		pmLabels,
	)

	portRcvErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_errors",
			Help: "port_rcv_errors",
		},
		pmLabels,
	)

	portXmitDiscard = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_discard",
			Help: "port_xmit_discard",
		},
		pmLabels,
	)

	portRcvSwitchRelayErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_switch_relay_errors",
			Help: "port_rcv_switch_relay_errors",
		},
		pmLabels,
	)

	excessiveBufferErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_excessive_buffer_errors",
			Help: "excessive_buffer_errors",
		},
		pmLabels,
	)

	localLinkIntegrityErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_local_link_integrity_errors",
			Help: "local_link_integrity_errors",
		},
		pmLabels,
	)

	portRcvConstraintErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_constraint_errors",
			Help: "port_rcv_constraint_errors",
		},
		pmLabels,
	)

	portXmitConstraintErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_constraint_errors",
			Help: "port_xmit_constraint_errors",
		},
		pmLabels,
	)

	vl15Dropped = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_vl15_dropped",
			Help: "vl15_dropped",
		},
		pmLabels,
	)
	portXmitData = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_data",
			Help: "port_xmit_data",
		},
		pmLabels,
	)
	portRcvData = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_data",
			Help: "port_rcv_data",
		},
		pmLabels,
	)
	portXmitPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_pkts",
			Help: "port_xmit_pkts",
		},
		pmLabels,
	)
	portRcvPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_pkts",
			Help: "port_rcv_pkts",
		},
		pmLabels,
	)
	portXmitWait = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_wait",
			Help: "port_xmit_wait",
		},
		pmLabels,
	)
	portXmitDataExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_data_extended",
			Help: "port_xmit_data_extended",
		},
		pmLabels,
	)
	portRcvDataExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_data_extended",
			Help: "port_rcv_data_extended",
		},
		pmLabels,
	)
	portXmitPktsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_pkts_extended",
			Help: "port_xmit_pkts_extended",
		},
		pmLabels,
	)
	portRcvPktsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_pkts_extended",
			Help: "port_rcv_pkts_extended",
		},
		pmLabels,
	)
	portUnicastXmitPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_unicast_xmit_pkts",
			Help: "port_unicast_xmit_pkts",
		},
		pmLabels,
	)
	portUnicastRcvPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_unicast_rcv_pkts",
			Help: "port_unicast_rcv_pkts",
		},
		pmLabels,
	)
	portMulticastXmitPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_multicast_xmit_pkts",
			Help: "port_multicast_xmit_pkts",
		},
		pmLabels,
	)
	portMulticastRcvPkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_multicast_rcv_pkts",
			Help: "port_multicast_rcv_pkts",
		},
		pmLabels,
	)
	symbolErrorCounterExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_symbol_error_counter_extended",
			Help: "symbol_error_counter_extended",
		},
		pmLabels,
	)
	linkErrorRecoveryCounterExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_error_recovery_counter_extended",
			Help: "link_error_recovery_counter_extended",
		},
		pmLabels,
	)
	linkDownedCounterExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_link_downed_counter_extended",
			Help: "link_downed_counter_extended",
		},
		pmLabels,
	)
	portRcvErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_errors_extended",
			Help: "port_rcv_errors_extended",
		},
		pmLabels,
	)
	portRcvRemotePhysicalErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_remote_physical_errors_extended",
			Help: "port_rcv_remote_physical_errors_extended",
		},
		pmLabels,
	)
	portRcvSwitchRelayErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_switch_relay_errors_extended",
			Help: "port_rcv_switch_relay_errors_extended",
		},
		pmLabels,
	)
	portXmitDiscardsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_discards_extended",
			Help: "port_xmit_discards_extended",
		},
		pmLabels,
	)
	portXmitConstraintErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_constraint_errors_extended",
			Help: "port_xmit_constraint_errors_extended",
		},
		pmLabels,
	)
	portRcvConstraintErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_rcv_constraint_errors_extended",
			Help: "port_rcv_constraint_errors_extended",
		},
		pmLabels,
	)
	localLinkIntegrityErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_local_link_integrity_errors_extended",
			Help: "local_link_integrity_errors_extended",
		},
		pmLabels,
	)
	excessiveBufferOverrunErrorsExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_excessive_buffer_overrun_errors_extended",
			Help: "excessive_buffer_overrun_errors_extended",
		},
		pmLabels,
	)
	vl15DroppedExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_vl15_dropped_extended",
			Help: "vl15_dropped_extended",
		},
		pmLabels,
	)
	portXmitWaitExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_xmit_wait_extended",
			Help: "port_xmit_wait_extended",
		},
		pmLabels,
	)
	qp1DroppedExtended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_qp1_dropped_extended",
			Help: "qp1_dropped_extended",
		},
		pmLabels,
	)
	retransmissionPerSec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_retransmission_per_sec",
			Help: "retransmission_per_sec",
		},
		pmLabels,
	)
	maxRetransmissionRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_max_retransmission_rate",
			Help: "max_retransmission_rate",
		},
		pmLabels,
	)
	portLocalPhysicalErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_local_physical_errors",
			Help: "port_local_physical_errors",
		},
		pmLabels,
	)
	portMalformedPacketErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_malformed_packet_errors",
			Help: "port_malformed_packet_errors",
		},
		pmLabels,
	)
	portBufferOverrunErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_buffer_overrun_errors",
			Help: "port_buffer_overrun_errors",
		},
		pmLabels,
	)
	portDlidMappingErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_dlid_mapping_errors",
			Help: "port_dlid_mapping_errors",
		},
		pmLabels,
	)
	portVlMappingErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_vl_mapping_errors",
			Help: "port_vl_mapping_errors",
		},
		pmLabels,
	)
	portLoopingErrors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_looping_errors",
			Help: "port_looping_errors",
		},
		pmLabels,
	)
	portInactiveDiscards = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_inactive_discards",
			Help: "port_inactive_discards",
		},
		pmLabels,
	)
	portNeighborMtuDiscards = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_neighbor_mtu_discards",
			Help: "port_neighbor_mtu_discards",
		},
		pmLabels,
	)
	portSwLifetimeLimitDiscards = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_sw_lifetime_limit_discards",
			Help: "port_sw_lifetime_limit_discards",
		},
		pmLabels,
	)
	portSwHoqLifetimeLimitDiscards = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "infiniband_port_sw_hoq_lifetime_limit_discards",
			Help: "port_sw_hoq_lifetime_limit_discards",
		},
		pmLabels,
	)
)

type Pmer interface {
	GetContent(filepath string) (*[]string, error)
	UpdateMetrics(blocks *[]string)
}

type LinkPm struct {
	FilePath string
}

type Pm struct {
	compoent string
	port     string
	lid      string
	guid     string
	device   string
	name     string
}

func init() {
	metrics := []prometheus.Collector{
		linkDownGauge,
		linkErrorRecoveryGauge,
		symbolErrorCounter,
		portRcvRemotePhysicalErrors,
		portRcvErrors,
		portXmitDiscard,
		portRcvSwitchRelayErrors,
		excessiveBufferErrors,
		localLinkIntegrityErrors,
		portRcvConstraintErrors,
		portXmitConstraintErrors,
		vl15Dropped,
		portXmitData,
		portRcvData,
		portXmitPkts,
		portRcvPkts,
		portXmitWait,
		portXmitDataExtended,
		portRcvDataExtended,
		portXmitPktsExtended,
		portRcvPktsExtended,
		portUnicastXmitPkts,
		portUnicastRcvPkts,
		portMulticastXmitPkts,
		portMulticastRcvPkts,
		symbolErrorCounterExtended,
		linkErrorRecoveryCounterExtended,
		linkDownedCounterExtended,
		portRcvErrorsExtended,
		portRcvRemotePhysicalErrorsExtended,
		portRcvSwitchRelayErrorsExtended,
		portXmitDiscardsExtended,
		portXmitConstraintErrorsExtended,
		portRcvConstraintErrorsExtended,
		localLinkIntegrityErrorsExtended,
		excessiveBufferOverrunErrorsExtended,
		vl15DroppedExtended,
		portXmitWaitExtended,
		qp1DroppedExtended,
		retransmissionPerSec,
		maxRetransmissionRate,
		portLocalPhysicalErrors,
		portMalformedPacketErrors,
		portBufferOverrunErrors,
		portDlidMappingErrors,
		portVlMappingErrors,
		portLoopingErrors,
		portInactiveDiscards,
		portNeighborMtuDiscards,
		portSwLifetimeLimitDiscards,
		portSwHoqLifetimeLimitDiscards,
	}
	for _, metric := range metrics {
		prometheus.MustRegister(metric)
	}
}
func (p *LinkPm) GetContent(filepath string) (*[]string, error) {
	fileContent, err := util.ReadFileContent(filepath)
	if err != nil {
		log.GetLogger().Error("read file error")
	}
	re := regexp.MustCompile(`(?m)Port=(\d+)\sLid=(\w+)\sGUID=(\w{18})\sDevice=(\d+)\sPort\sName=(.*)`)
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

func (p *LinkPm) UpdateMetrics(blocks *[]string) {
	for _, block := range *blocks {
		switchCaExpr := `Port=(\d+)\sLid=(\w+)\sGUID=(\w{18})\sDevice=(\d+)\sPort\sName=(.*)`
		switchCaMatch, err := regexp.Compile(switchCaExpr)
		if err != nil {
			log.GetLogger().Error("switch or ca error compiling regex")
			break
		}
		subSwitchCaMatch := switchCaMatch.FindStringSubmatch(block)
		pm := Pm{
			compoent: global.COMPONENT_SW,
			port:     subSwitchCaMatch[1],
			lid:      subSwitchCaMatch[2],
			guid:     subSwitchCaMatch[3],
			device:   subSwitchCaMatch[4],
			name:     subSwitchCaMatch[5],
		}
		fmt.Println(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name)
		getValue := func(repr string) (value float64) {
			re := regexp.MustCompile(repr)
			match := re.FindStringSubmatch(block)
			dec, err := strconv.ParseInt(match[1], 0, 64)
			fmt.Println(err)
			if err != nil {
				value = 0
				return value
			}
			value = float64(dec)
			return value
		}
		linkDownGauge.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(getValue(`link_down_gauge=(\w+)`))
		linkErrorRecoveryGauge.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		symbolErrorCounter.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvRemotePhysicalErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitDiscard.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvSwitchRelayErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		excessiveBufferErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		localLinkIntegrityErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvConstraintErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitConstraintErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		vl15Dropped.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitData.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvData.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitWait.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitDataExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvDataExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitPktsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvPktsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portUnicastXmitPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portUnicastRcvPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portMulticastXmitPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portMulticastRcvPkts.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		symbolErrorCounterExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		linkErrorRecoveryCounterExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		linkDownedCounterExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvRemotePhysicalErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvSwitchRelayErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitDiscardsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitConstraintErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portRcvConstraintErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		localLinkIntegrityErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		excessiveBufferOverrunErrorsExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		vl15DroppedExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portXmitWaitExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		qp1DroppedExtended.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		retransmissionPerSec.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		maxRetransmissionRate.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portLocalPhysicalErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portMalformedPacketErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portBufferOverrunErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portDlidMappingErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portVlMappingErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portLoopingErrors.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portInactiveDiscards.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portNeighborMtuDiscards.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portSwLifetimeLimitDiscards.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
		portSwHoqLifetimeLimitDiscards.
			WithLabelValues(pm.compoent, pm.port, pm.lid, pm.guid, pm.device, pm.name).
			Set(0)
	}
}
