package updateTracker

import (
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Juby210/gplayapi-go"
	"github.com/Juby210/gplayapi-go/gpproto"
)

const (
	arm64 = "arm64_v8a"
	arm32 = "armeabi_v7a"
	x64   = "x86_64"
	x86   = "x86"

	DefaultArch = arm64
)

var (
	MissingArchSplits = []string{"config." + arm32, "config." + x64, "config." + x86}
	MissingDpiSplits  = []string{"config.hdpi"}
)

type GooglePlayChecker struct {
	clients map[string]*gplayapi.GooglePlayClient

	AccountConfig common.GooglePlayChannelConfig
	Channel       string
}

func (c *GooglePlayChecker) init(arch string) (err error) {
	if c.clients[arch] == nil {
		sFile := c.getSessionFile(arch)

		var device *gplayapi.DeviceInfo
		switch arch {
		case arm32:
			device = gplayapi.Redmi4
		case x64:
			device = gplayapi.Emulator_x86_64
		case x86:
			device = gplayapi.Emulator_x86
		default:
			device = gplayapi.Pixel3a
		}

		c.clients[arch], err = gplayapi.LoadSessionWithDeviceInfo(sFile, device)
		if err != nil {
			c.clients[arch], err = gplayapi.NewClientWithDeviceInfo(
				c.AccountConfig.Email, c.AccountConfig.AASToken, device)
			if err == nil {
				c.clients[arch].SaveSession(sFile)
			}
		}
		if c.clients[arch] != nil {
			c.clients[arch].SessionFile = sFile
		}
	}
	return
}

func (c *GooglePlayChecker) getSessionFile(arch string) string {
	return "_sessions/session" + common.ToTitle(c.Channel) + "." + arch + ".json"
}

func (c *GooglePlayChecker) Check() (v int, app *gplayapi.App, err error) {
	err = c.init(DefaultArch)
	if err != nil {
		return
	}
	app, err = c.clients[DefaultArch].GetAppDetails(discordPkg)
	if err == nil {
		v = app.VersionCode
	}
	return
}

func (c *GooglePlayChecker) GetDownloadData(version int, arch string) (data *gpproto.AndroidAppDeliveryData, err error) {
	err = c.init(arch)
	if err != nil {
		return
	}
	data, err = c.clients[arch].Purchase(discordPkg, version)
	return
}

var gpCheckers = map[string]*GooglePlayChecker{}

func initGPCheckers(cfg map[string]common.GooglePlayChannelConfig) {
	for channel, accountConfig := range cfg {
		gpCheckers[channel] = &GooglePlayChecker{
			clients:       map[string]*gplayapi.GooglePlayClient{},
			AccountConfig: accountConfig,
			Channel:       channel,
		}
	}
}
