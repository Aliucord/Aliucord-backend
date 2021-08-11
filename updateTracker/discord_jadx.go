package updateTracker

import (
	"strconv"
	"strings"

	"github.com/Juby210/admh"
	"github.com/Juby210/gplayapi-go"
)

func Extract(channel string, app *gplayapi.App) error {
	logger.Printf("Running admh for discord release channel %s (%d)", channel, app.VersionCode)
	cfg := config.UpdateTracker.DiscordJADX
	repo := cfg.RepoBase + strings.Title(channel)
	err := admh.DownloadAndExtractAPK(
		discordPkg,
		app.VersionCode,
		cfg.WorkDir,
		repo,
		admh.GetDefaultJadxFlags(),
		DownloadUrlResolver,
	)
	if err != nil {
		return err
	}
	if cfg.AutoPush {
		return admh.Push(repo, app.VersionName, strconv.Itoa(app.VersionCode))
	}
	return nil
}

func DownloadUrlResolver(_ string, version int) (url string, err error) {
	return GetDownloadURL(version, true)
}
