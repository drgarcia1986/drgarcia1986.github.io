package influxdb

type Client interface {
	Write(class, title string) error
}
