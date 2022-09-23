package updateTracker

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
	dlCache = map[string]map[int]*DlCache{
		arm64: {},
		arm32: {},
		x64:   {},
		x86:   {},
	}
	wh *webhook.Client

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
	if !gpChecker.AccountConfig.Webhook {
		return
	}

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

	for _, cache := range dlCache {
		if dl, ok := cache[data.Version]; ok && dl.Error != nil {
			delete(cache, data.Version)
		}
	}

	if wh == nil {
		if data.Version < gpVersion {
			data.Version = gpVersion
			logger.LogIfErr(cache.Persist())
		}
		return
	}

	if update {
		dl, err := GetDownloadData(gpVersion, DefaultArch, true)
		if err == nil {
			var description string
			if len(dl.Splits) > 0 {
				description += fmt.Sprintf("[base.apk](%s/download/discord?v=%d)", config.Origin, gpVersion)

				var archSplits []string
				var languageSplits []string
				dpiSplits := MissingDpiSplits

				for splitName := range dl.Splits {
					if splitName == "config."+DefaultArch {
						archSplits = append(archSplits, splitName)
					} else if strings.Contains(splitName, "dpi") {
						dpiSplits = append(dpiSplits, splitName)
					} else {
						languageSplits = append(languageSplits, splitName)
					}
				}

				archSplits = append(archSplits, MissingArchSplits...)

				format := fmt.Sprintf("\n[%%s.apk](%s/download/discord?v=%d&split=%%s)", config.Origin, gpVersion)
				description += joinSplits(archSplits, "Architecture", format)
				description += joinSplits(dpiSplits, "DPI", format)
				description += joinSplits(languageSplits, "Language", format)
			} else {
				description = fmt.Sprintf("[Click here to download apk](%s/download/discord?v=%d)", config.Origin, gpVersion)
			}
			err = wh.Execute(webhook.ExecuteData{
				Username: "Update - " + common.ToTitle(channel),
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

func joinSplits(splits []string, title, format string) string {
	ret := "\n\n**" + title + " splits**:"
	for _, splitName := range splits {
		ret += fmt.Sprintf(format, splitName, splitName)
	}
	return ret
}

func GetDownloadData(version int, arch string, bypass bool) (dl *DlCache, err error) {
	dl, ok := dlCache[arch][version]
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

	data, err := gpCheckers["alpha"].GetDownloadData(version, arch)
	if dl == nil {
		dl = &DlCache{GP: true}
		dlCache[arch][version] = dl
	}
	dl.Expiry = time.Now().Unix() + 22*60*60
	if err != nil {
		dl.Error = err
		return
	}
	if data.SplitDeliveryData == nil {
		dl.URL = data.GetDownloadUrl()
	} else {
		splits := map[string]string{}
		for _, split := range data.SplitDeliveryData {
			splits[split.GetName()] = split.GetDownloadUrl()
		}
		dl.URL = data.GetDownloadUrl()
		dl.Splits = splits
	}
	return
}

func GetDownloadURL(version int, split string, bypass bool) (url string, err error) {
	if !bypass {
		apkName := "/" + strconv.Itoa(version) + "/"
		if split == "" {
			apkName += "base.apk"
		} else {
			apkName += split + ".apk"
		}
		if _, err = os.Stat(config.ApkCacheDir + apkName); err == nil {
			return config.Origin + "/download/direct" + apkName, nil
		}
	}

	var arch string
	switch split {
	case "config." + arm32, "config.hdpi":
		arch = arm32
	case "config." + x64:
		arch = x64
	case "config." + x86:
		arch = x86
	default:
		arch = arm64
	}

	dl, err := GetDownloadData(version, arch, bypass)
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
