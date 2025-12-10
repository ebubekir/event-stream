package domain

import "time"

type ChannelType string

const (
	ChannelTypeWeb     ChannelType = "web"
	ChannelTypeMobile  ChannelType = "mobile"
	ChannelTypeDesktop ChannelType = "desktop"
	ChannelTypeTV      ChannelType = "tv"
	ChannelTypeConsole ChannelType = "console"
	ChannelTypeOther   ChannelType = "other"
)

type Param struct {
	Key          string
	StringValue  string
	NumberValue  float64
	BooleanValue bool
}

type Device struct {
	Category               string
	MobileBrandName        string
	MobileModelName        string
	OperatingSystem        string
	OperatingSystemVersion string
	Language               string
	BrowserName            string
	BrowserVersion         string
	Hostname               string
}

type Geo struct {
	Continent    string
	SubContinent string
	Country      string
	Region       string
	Metro        string
	City         string
}

type AppInfo struct {
	ID      string
	Version string
}

type Item struct {
	ID            string
	Name          string
	Brand         string
	Variant       string
	PriceInUsd    float64
	Quantity      int
	RevenueInUsd  float64
	LocationId    string
	ListId        string
	ListName      string
	PromotionId   string
	PromotionName string
	Params        []Param
}

type Event struct {
	ID                string
	Timestamp         uint16
	PreviousTimestamp uint16
	Date              time.Time
	Name              string
	ChannelType
	EventParams  []Param
	UserID       string
	UserPseudoID string
	UserParams   []Param
	Device       Device
	AppInfo      AppInfo
	Items        []Item
}
