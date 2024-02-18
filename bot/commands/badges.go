package commands

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type (
	CustomBadge struct {
		Id   string `json:"id,omitempty"`
		Url  string `json:"url,omitempty"`
		Text string `json:"text"`
	}
	UserBadges struct {
		Roles  []string      `json:"roles,omitempty"`
		Custom []CustomBadge `json:"custom,omitempty"`
	}
	BadgeData struct {
		Guilds map[string]CustomBadge `json:"guilds"`
		Users  map[string]*UserBadges `json:"users"`
	}
)

var (
	Badges BadgeData

	basePath = "static/files/badges"
	dataPath = basePath + "/data.json"
)

func init() {
	user := &discord.UserOption{
		OptionName: "user", Description: "user",
	}
	guild := &discord.StringOption{
		OptionName: "guild-id", Description: "guild id",
	}
	text := &discord.StringOption{
		OptionName: "text", Description: "text",
	}
	imageUrl := &discord.StringOption{
		OptionName: "image-url", Description: "image url",
	}
	role := &discord.StringOption{
		OptionName: "role", Description: "role (user only)",
	}
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "badges",
			Description: "Manage badges",
			Options: []discord.CommandOption{
				&discord.SubcommandOption{
					OptionName:  "add",
					Description: "Add badge to user/guild",
					Options: []discord.CommandOptionValue{
						user, guild, text, imageUrl, role,
					},
				},
				// TODO: add edit
				// &discord.SubcommandOption{
				// 	OptionName:  "edit",
				// 	Description: "Edit badge",
				// 	Options: []discord.CommandOptionValue{
				// 		user, guild,
				// 		&discord.IntegerOption{
				// 			OptionName:  "index",
				// 			Description: "index (required for user badge)",
				// 		},
				// 		text, imageUrl,
				// 	},
				// },
				&discord.SubcommandOption{
					OptionName:  "remove",
					Description: "Remove badge(s)",
					Options: []discord.CommandOptionValue{
						user, guild,
						&discord.IntegerOption{
							OptionName:  "index",
							Description: "index (user only, to delete only one)",
						},
						role,
					},
				},
			},
		},
		OwnerOnly: true,
		Execute:   badgesCommand,
	})

	if _, err := os.Stat(dataPath); err == nil {
		f, err := os.Open(dataPath)
		if err == nil {
			if err = json.NewDecoder(f).Decode(&Badges); err != nil {
				logger.LogWithCtxIfErr("loading badge data", err)
				Badges = BadgeData{Guilds: map[string]CustomBadge{}, Users: map[string]*UserBadges{}}
			}
			f.Close()
		} else {
			logger.LogWithCtxIfErr("loading badge data", err)
			Badges = BadgeData{Guilds: map[string]CustomBadge{}, Users: map[string]*UserBadges{}}
		}
	} else {
		Badges = BadgeData{Guilds: map[string]CustomBadge{}, Users: map[string]*UserBadges{}}
	}
}

func saveBadges() {
	if _, err := os.Stat(basePath); err != nil {
		os.MkdirAll(basePath, 0777)
	}
	f, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err == nil {
		if err = json.NewEncoder(f).Encode(&Badges); err != nil {
			logger.LogWithCtxIfErr("saving badges", err)
		}
		f.Close()
	} else {
		logger.LogWithCtxIfErr("saving badges", err)
	}
}

func badgesCommand(ev *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	cmd := d.Options[0]
	subcommand := cmd.Name
	opts := cmd.Options
	user := findOption2(opts, "user")
	guild := findOption2(opts, "guild-id")
	if user == nil && guild == nil {
		return ephemeralReply(ev, "No user or guild specified")
	}
	switch subcommand {
	case "add":
		role := findOption2(opts, "role")
		if role != nil {
			if user == nil {
				return ephemeralReply(ev, "No user specified for role")
			}

			id := user.String()
			badges, ok := Badges.Users[id]
			if !ok {
				badges = &UserBadges{}
				Badges.Users[id] = badges
			}
			badges.Roles = append(badges.Roles, role.String())
		}

		url := findOption2(opts, "image-url")
		if url == nil {
			if role == nil {
				return ephemeralReply(ev, "Missing required badge data")
			} else {
				saveBadges()
				return ephemeralReply(ev, "Added")
			}
		}

		text := findOption2(opts, "text")
		if text == nil {
			return ephemeralReply(ev, "Text is required")
		}

		imageUrl := url.String()
		resp, err := http.Get(imageUrl)
		if err != nil {
			return ephemeralReply(ev, err.Error())
		}
		defer resp.Body.Close()

		imgData, err := io.ReadAll(resp.Body)
		if err != nil {
			return ephemeralReply(ev, err.Error())
		}

		h := sha1.New()
		h.Write(imgData)
		hash := hex.EncodeToString(h.Sum(nil))

		split := strings.Split(strings.Split(imageUrl, "?")[0], ".")
		ext := split[len(split)-1]
		fileName := "/" + hash + "." + ext

		if user != nil {
			id := user.String()

			path := "/users/" + id
			fPath := basePath + path
			if _, err = os.Stat(fPath); err != nil {
				err = os.MkdirAll(fPath, 0777)
				if err != nil {
					return ephemeralReply(ev, err.Error())
				}
			}
			if err = os.WriteFile(fPath+fileName, imgData, 0644); err != nil {
				return ephemeralReply(ev, err.Error())
			}

			badges, ok := Badges.Users[id]
			if !ok {
				badges = &UserBadges{}
				Badges.Users[id] = badges
			}
			badges.Custom = append(badges.Custom, CustomBadge{
				Url:  config.Origin + "/files/badges" + path + fileName,
				Text: text.String(),
			})
		}

		if guild != nil {
			id := guild.String()

			path := "/guilds/" + id
			fPath := basePath + path
			if _, err = os.Stat(fPath); err != nil {
				err = os.MkdirAll(fPath, 0777)
				if err != nil {
					return ephemeralReply(ev, err.Error())
				}
			}
			if err = os.WriteFile(fPath+fileName, imgData, 0644); err != nil {
				return ephemeralReply(ev, err.Error())
			}

			Badges.Guilds[id] = CustomBadge{
				Url:  config.Origin + "/files/badges" + path + fileName,
				Text: text.String(),
			}
		}

		saveBadges()
		return ephemeralReply(ev, "Added")
	case "remove":
		if user != nil {
			id := user.String()
			if badges, ok := Badges.Users[id]; ok {
				index := findOption2(opts, "index")
				role := findOption2(opts, "role")
				if index == nil && role == nil {
					delete(Badges.Users, id)
					os.RemoveAll(basePath + "/users/" + id)
				} else {
					if index != nil {
						idx, err := index.IntValue()
						if err != nil {
							return ephemeralReply(ev, err.Error())
						}
						idxInt := int(idx)
						if len(badges.Custom) > idxInt {
							if len(badges.Custom) == 1 {
								if len(badges.Roles) > 0 {
									badges.Custom = []CustomBadge{}
								} else {
									delete(Badges.Users, id)
								}
								os.RemoveAll(basePath + "/users/" + id)
							} else {
								path := strings.TrimPrefix(badges.Custom[idxInt].Url, config.Origin+"/files/badges")
								badges.Custom = append(badges.Custom[:idxInt], badges.Custom[idxInt+1:]...)
								os.Remove(basePath + path)
							}
						}
					}
					if role != nil {
						rol := role.String()
						i := 0
						for _, r := range badges.Roles {
							if r != rol {
								badges.Roles[i] = r
								i++
							}
						}
						badges.Roles = badges.Roles[:i]
					}
				}
			} else {
				return ephemeralReply(ev, "This user doesn't have any badges")
			}
		}

		if guild != nil {
			id := guild.String()
			delete(Badges.Guilds, id)
			os.RemoveAll(basePath + "/guilds/" + id)
		}

		saveBadges()
		return ephemeralReply(ev, "Removed")
	}
	return ephemeralReply(ev, "No such subcommand: "+subcommand)
}
