package updateTracker

import (
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Juby210/gplayapi-go"
	"github.com/Juby210/gplayapi-go/gpproto"
)

type GooglePlayChecker struct {
	client *gplayapi.GooglePlayClient

	AccountConfig common.GooglePlayChannelConfig
	Channel       string
}

func (c *GooglePlayChecker) init() (err error) {
	if c.client == nil {
		sFile := c.getSessionFile()
		c.client, err = gplayapi.LoadSession(sFile)
		if err != nil {
			c.client, err = gplayapi.NewClient(c.AccountConfig.Email, c.AccountConfig.AASToken)
			if err == nil {
				c.client.SaveSession(sFile)
			}
		}
		if c.client != nil {
			c.client.SessionFile = sFile
		}
	}
	return
}

func (c *GooglePlayChecker) getSessionFile() string {
	return "_session" + strings.Title(c.Channel) + ".json"
}

func (c *GooglePlayChecker) Check() (v int, app *gplayapi.App, err error) {
	err = c.init()
	if err != nil {
		return
	}
	app, err = c.client.GetAppDetails(discordPkg)
	if err == nil {
		v = app.VersionCode
	}
	return
}

func (c *GooglePlayChecker) GetDownloadData(version int) (data *gpproto.AndroidAppDeliveryData, err error) {
	err = c.init()
	if err != nil {
		return
	}
	data, err = c.client.Purchase(discordPkg, version)
	return
}

var gpCheckers = map[string]*GooglePlayChecker{}

func initGPCheckers(cfg map[string]common.GooglePlayChannelConfig) {
	for channel, accountConfig := range cfg {
		gpCheckers[channel] = &GooglePlayChecker{AccountConfig: accountConfig, Channel: channel}
	}
}
