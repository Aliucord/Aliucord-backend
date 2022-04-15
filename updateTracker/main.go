package updateTracker

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
)

type DlCache struct {
	URL    string
	Splits map[string]string
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
	dlCache = map[int]*DlCache{}
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

	for channel := range config.UpdateTracker.GooglePlay {
		check(channel)
	}
}

func check(channel string) {
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
		logger.LogIfErr(cache.Persist())
		return
	}
	update := data.Version < gpVersion

	if dl, ok := dlCache[data.Version]; ok && dl.Error != nil {
		delete(dlCache, data.Version)
	}

	if wh == nil {
		if data.Version < gpVersion {
			data.Version = gpVersion
			logger.LogIfErr(cache.Persist())
		}
		return
	}

	if update {
		dl, err := getDownloadData(gpVersion, true)
		if err == nil {
			var description string
			if len(dl.Splits) > 0 {
				description += fmt.Sprintf("[base.apk](%s/download/discord?v=%d)", config.Origin, gpVersion)
				for splitName := range dl.Splits {
					description += fmt.Sprintf("\n[%s.apk](%s/download/discord?v=%d&split=%s)", splitName, config.Origin, gpVersion, splitName)
				}
			} else {
				description = fmt.Sprintf("[Click here to download apk](%s/download/discord?v=%d)", config.Origin, gpVersion)
			}
			err = wh.Execute(webhook.ExecuteData{
				Username: "Discord Update - " + strings.Title(channel),
				Embeds: []discord.Embed{{
					Author: &discord.EmbedAuthor{
						Name: app.DisplayName,
						Icon: app.IconImage.GetImageUrl(),
						URL:  googlePlayURL,
					},
					Title:       fmt.Sprintf("New version: **%s (%d)**", app.VersionName, gpVersion),
					Description: description,
					Color:       7506394,
				}},
			})
			logger.LogIfErr(err)
		} else {
			logger.Println(err)
		}
	}

	if data.Version < gpVersion {
		data.Version = gpVersion
		logger.LogIfErr(cache.Persist())
	}
}

func getDownloadData(version int, bypass bool) (dl *DlCache, err error) {
	dl, ok := dlCache[version]
	if ok && (!dl.GP || dl.Expiry > time.Now().Unix()) {
		return dl, dl.Error
	}

	if !bypass {
		if version < config.MinDownloadVer || config.MaxDownloadVer != 0 && config.MaxDownloadVer < version {
			return nil, errors.New("you are trying to request too old or new version")
		}
		// gets release channel from version code, max release channel is 2 so all versions with higher type are invalid
		if version/100%10 > 2 {
			return nil, errors.New("invalid version code")
		}
	}

	data, err := gpCheckers["alpha"].GetDownloadData(version)
	if err != nil {
		dl = &DlCache{URL: "", GP: true, Expiry: time.Now().Unix() + 22*60*60, Error: err}
		dlCache[version] = dl
		return
	}
	if data.SplitDeliveryData == nil {
		dl = &DlCache{URL: data.GetDownloadUrl(), GP: true, Expiry: time.Now().Unix() + 22*60*60}
	} else {
		splits := map[string]string{}
		for _, split := range data.SplitDeliveryData {
			splits[split.GetName()] = split.GetDownloadUrl()
		}
		dl = &DlCache{
			URL:    data.GetDownloadUrl(),
			Splits: splits,
			GP:     true,
			Expiry: time.Now().Unix() + 22*60*60,
		}
	}
	dlCache[version] = dl
	return
}

func GetDownloadURL(version int, split string, bypass bool) (url string, err error) {
	// if !bypass {
	// 	apkName := "com.discord-" + strconv.Itoa(version) + ".apk"
	// 	if _, err = os.Stat(config.UpdateTracker.DiscordJADX.WorkDir + "/apk/" + apkName); err == nil {
	// 		return config.Origin + "/download/direct/" + apkName, nil
	// 	}
	// }

	if url, ok := config.Mirrors[version]; ok {
		return url, nil
	}

	dl, err := getDownloadData(version, bypass)
	if err != nil {
		return
	}
	if split != "" {
		if splitUrl, ok := dl.Splits[split]; ok {
			url = splitUrl
		} else {
			err = errors.New("split not found")
		}
	} else {
		url = dl.URL
	}

	return
}
