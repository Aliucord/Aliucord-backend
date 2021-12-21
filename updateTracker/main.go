package updateTracker

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Juby210/admh/aptoide"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
)

type DlCache struct {
	URL    string
	GP     bool
	Expiry int64
	Error  error
}

const (
	discordPkg    = "com.discord"
	googlePlayURL = "https://play.google.com/store/apps/details?id=" + discordPkg
)

var (
	config  *common.Config
	dlCache = map[int]DlCache{}
	wh      *webhook.Client

	logger = common.NewLogger("[updateTracker]")
)

func StartUpdateTracker(cfg *common.Config) {
	config = cfg
	c := cfg.UpdateTracker
	initGPCheckers(c.GooglePlay)

	if !c.Enabled {
		return
	}

	var err error
	cache, err = newCache(c.Cache)
	if err != nil {
		logger.Println("load cache error", err)
		return
	}
	if c.Webhook.Enabled {
		wh = webhook.New(c.Webhook.ID, c.Webhook.Token)
	}
	go func() {
		for {
			Check()
			time.Sleep(15 * time.Minute)
		}
	}()
}

func Check() {
	logger.Println("Checking for discord updates")

	for channel, cfg := range config.UpdateTracker.GooglePlay {
		check(channel, cfg)
	}
}

func check(channel string, cfg common.GooglePlayChannelConfig) {
	gpChecker := gpCheckers[channel]

	gpVersion, app, err := gpChecker.Check()
	if err != nil {
		logger.Println("Failed to check", err)
		return
	}

	var data *CacheData
	if cData, ok := cache.Data[channel]; ok {
		data = cData
	} else {
		data = &CacheData{}
		cache.Data[channel] = data
	}

	if config.UpdateTracker.IgnoreFirstUpdate && data.Version == 0 {
		data.Version = gpVersion
		data.JADX = true
		logger.LogIfErr(cache.Persist())
		return
	}
	update := data.Version < gpVersion

	if config.UpdateTracker.DiscordJADX.Enabled && cfg.JADX {
		if update || !data.JADX {
			err = Extract(channel, app)
			logger.LogIfErr(err)
			if err == nil {
				data.Version = gpVersion
				data.JADX = true
				logger.LogIfErr(cache.Persist())
			}
		}
	}

	if dl, ok := dlCache[data.Version]; ok && dl.Error != nil {
		delete(dlCache, data.Version)
	}

	if wh == nil {
		if data.Version < gpVersion {
			data.Version = gpVersion
			data.JADX = false
			logger.LogIfErr(cache.Persist())
		}
		return
	}

	if update {
		err = wh.Execute(webhook.ExecuteData{
			Username: "Discord Update - " + strings.Title(channel),
			Embeds: []discord.Embed{{
				Author: &discord.EmbedAuthor{
					Name: app.DisplayName,
					Icon: app.IconImage.GetImageUrl(),
					URL:  googlePlayURL,
				},
				Title:       fmt.Sprintf("New version: **%s (%d)**", app.VersionName, gpVersion),
				Description: fmt.Sprintf("[Click here to download apk](%s/download/discord?v=%d)", config.Origin, gpVersion),
				Color:       7506394,
			}},
		})
		logger.LogIfErr(err)
	}

	if data.Version < gpVersion {
		data.Version = gpVersion
		data.JADX = false
		logger.LogIfErr(cache.Persist())
	}
}

func GetDownloadURL(version int, bypass bool) (url string, err error) {
	if !bypass {
		apkName := "com.discord-" + strconv.Itoa(version) + ".apk"
		if _, err = os.Stat(config.UpdateTracker.DiscordJADX.WorkDir + "/apk/" + apkName); err == nil {
			return config.Origin + "/download/direct/" + apkName, nil
		}
	}

	if url, ok := config.Mirrors[version]; ok {
		return url, nil
	}

	dl, ok := dlCache[version]
	if ok && (!dl.GP || dl.Expiry > time.Now().Unix()) {
		return dl.URL, dl.Error
	}

	if !bypass {
		if version < config.MinDownloadVer || config.MaxDownloadVer != 0 && config.MaxDownloadVer < version {
			return "", errors.New("you are trying to request too old or new version")
		}
		// gets release channel from version code, max release channel is 2 so all versions with higher type are invalid
		if version/100%10 > 2 {
			return "", errors.New("invalid version code")
		}
	}

	useAptoide := true
	for _, v := range config.DisableAptoide {
		if v == version {
			useAptoide = false
			break
		}
	}

	if useAptoide {
		url, err = aptoide.DownloadUrlResolver(discordPkg, version)
		if err == nil {
			dlCache[version] = DlCache{URL: url}
			return
		}
	}

	url, err = gpCheckers["alpha"].GetDownloadURL(version)
	dlCache[version] = DlCache{URL: url, GP: true, Expiry: time.Now().Unix() + 22*60*60, Error: err}
	return
}
