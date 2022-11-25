package commands

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Aliucord/Aliucord-backend/updateTracker"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	options := []discord.CommandOptionValue{
		&discord.IntegerOption{
			OptionName:  "version",
			Description: "Version",
			Required:    true,
			Min:         option.NewInt(80000),
			Max:         option.NewInt(200000),
		},
	}
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "apkcache",
			Description: "Manage apk cache",
			Options: []discord.CommandOption{
				&discord.SubcommandOption{
					OptionName:  "list",
					Description: "List apks",
				},
				&discord.SubcommandOption{
					OptionName:  "cache",
					Description: "Cache apks",
					Options:     options,
				},
				&discord.SubcommandOption{
					OptionName:  "purge",
					Description: "Purge apks",
					Options:     options,
				},
			},
		},
		OwnerOnly: true,
		Execute:   apkCacheCommand,
	})
}

func apkCacheCommand(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	cmd := d.Options[0]
	subcommand := cmd.Name
	switch subcommand {
	case "list":
		files, err := os.ReadDir(config.ApkCacheDir)
		if err != nil {
			return ephemeralReply(ev, "Failed to read apk cache dir: "+err.Error())
		}
		var sb strings.Builder
		for _, version := range files {
			if version.IsDir() {
				sb.WriteString("- **")
				ver := version.Name()
				sb.WriteString(ver)
				sb.WriteString("** (")
				verFiles, err := os.ReadDir(config.ApkCacheDir + "/" + ver)
				if err == nil {
					sb.WriteString(strings.Join(common.SliceTransform(verFiles, func(file os.DirEntry) string {
						return strings.TrimSuffix(file.Name(), ".apk")
					}), " "))
				} else {
					sb.WriteString(err.Error())
				}
				sb.WriteString(")\n")
			}
		}
		ret := "Cached apks:"
		if sb.Len() == 0 {
			ret += " none"
		} else {
			ret += "\n" + sb.String()
		}
		return ephemeralReply(ev, ret)

	case "cache":
		version, err := cmd.Options[0].IntValue()
		if err != nil {
			return ephemeralReply(ev, err.Error())
		}
		ver := int(version)
		versionStr := strconv.Itoa(ver)

		basePath := config.ApkCacheDir + "/" + versionStr + "/"
		if _, err = os.Stat(basePath); err == nil {
			return ephemeralReply(ev, "This version is already cached, purge cache for this version first")
		}

		err = s.RespondInteraction(ev.ID, ev.Token, api.InteractionResponse{
			Type: api.DeferredMessageInteractionWithSource,
			Data: &api.InteractionResponseData{Flags: discord.EphemeralMessage},
		})
		if err != nil {
			return err
		}

		dl, err := updateTracker.GetDownloadData(ver, updateTracker.DefaultArch, true)
		if err != nil {
			return editReply(ev, err.Error())
		}

		basePath2 := config.ApkCacheDir + "/" + versionStr + ".tmp/"
		os.MkdirAll(basePath2, 0777)

		handleFail := func(name string, err error) error {
			os.RemoveAll(basePath2)
			return editReply(ev, "Failed to download "+name+".apk: "+err.Error())
		}

		if err = downloadApk(dl.URL, basePath2+"base.apk"); err != nil {
			return handleFail("base", err)
		}

		cached := []string{"base"}
		if len(dl.Splits) > 0 {
			for split, url := range dl.Splits {
				if err = downloadApk(url, basePath2+split+".apk"); err != nil {
					return handleFail(split, err)
				}
				cached = append(cached, split)
			}

			for _, split := range append(updateTracker.MissingArchSplits, updateTracker.MissingDpiSplits...) {
				url, err := updateTracker.GetDownloadURL(ver, split, true)
				if err != nil {
					return handleFail(split, err)
				}

				if err = downloadApk(url, basePath2+split+".apk"); err != nil {
					return handleFail(split, err)
				}
				cached = append(cached, split)
			}
		}

		os.RemoveAll(basePath)
		if err = os.Rename(basePath2, basePath); err != nil {
			return editReply(ev, "Failed to rename temp dir to normal: "+err.Error())
		}

		return editReply(ev, "Done! Cached "+strings.Join(cached, " ")+" for **"+versionStr+"** version")

	case "purge":
		ver := cmd.Options[0].String()
		path := config.ApkCacheDir + "/" + ver
		if _, err := os.Stat(path); err == nil {
			err = os.RemoveAll(path)
			if err == nil {
				return ephemeralReply(ev, "Purged cache for selected version")
			} else {
				return ephemeralReply(ev, "Failed to purge cache: "+err.Error())
			}
		}
		return ephemeralReply(ev, "This version isn't cached")
	}
	return ephemeralReply(ev, "No such subcommand: "+subcommand)
}

func downloadApk(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return file.Close()
}
