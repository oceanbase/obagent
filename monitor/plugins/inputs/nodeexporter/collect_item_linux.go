package nodeexporter

const (
	ntp              = "collector.ntp"
	arp              = "collector.arp"
	bcache           = "collector.bcache"
	bonding          = "collector.bonding"
	btrfs            = "collector.btrfs"
	conntrack        = "collector.conntrack"
	cpufreq          = "collector.cpufreq"
	edac             = "collector.edac"
	entropy          = "collector.entropy"
	fibrechannel     = "collector.fibrechannel"
	hwmon            = "collector.hwmon"
	infiniband       = "collector.infiniband"
	ipvs             = "collector.ipvs"
	mdadm            = "collector.mdadm"
	netclass         = "collector.netclass"
	nfs              = "collector.nfs"
	nfsd             = "collector.nfsd"
	nvme             = "collector.nvme"
	powersupplyclass = "collector.powersupplyclass"
	pressure         = "collector.pressure"
	rapl             = "collector.rapl"
	schedstat        = "collector.schedstat"
	softnet          = "collector.softnet"
	stat             = "collector.stat"
	tapestats        = "collector.tapestats"
	textfile         = "collector.textfile"
	thermal_zone     = "collector.thermal_zone"
	timec            = "collector.time"
	timex            = "collector.timex"
	udp_queues       = "collector.udp_queues"
	uname            = "collector.uname"
	vmstat           = "collector.vmstat"
	xfs              = "collector.xfs"
	zfs              = "collector.zfs"
)

var collectItems = []string{
	ntp,
}

var noCollectItems = []string{
	arp,
	bcache,
	bonding,
	btrfs,
	conntrack,
	cpufreq,
	edac,
	entropy,
	fibrechannel,
	hwmon,
	infiniband,
	ipvs,
	mdadm,
	netclass,
	nfs,
	nfsd,
	nvme,
	powersupplyclass,
	pressure,
	rapl,
	schedstat,
	softnet,
	stat,
	tapestats,
	textfile,
	thermal_zone,
	timec,
	timex,
	udp_queues,
	uname,
	vmstat,
	xfs,
	zfs,
}
