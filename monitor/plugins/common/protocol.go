package common

type Protocol string

const (
	Prometheus Protocol = "prometheus"
	PromeProto Protocol = "promeproto"
	Json       Protocol = "json"
	Csv        Protocol = "csv"
	Influxdb   Protocol = "influxdb"
)

type TimestampPrecision string

const (
	Second      TimestampPrecision = "second"
	Millisecond TimestampPrecision = "millisecond"
)
