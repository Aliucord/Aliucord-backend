package commands

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Aliucord/Aliucord-backend/database"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

var (
	mentionRegex  = regexp.MustCompile("<@!?\\d{17,19}>")
	idRegex       = regexp.MustCompile("\\d{17,19}")
	durationRegex = regexp.MustCompile("(\\d+)(s|seconds|secs?|m|mins?|minutes|h|hours|d|days|w|weeks)")
)

type nameAndID struct {
	name string
	id   discord.RoleID
}

func initModCommands() {
	roles := []nameAndID{
		{name: "supportmute", id: config.RoleIDs.SupportMuted},
		{name: "devmute", id: config.RoleIDs.DevMuted},
		{name: "prdmute", id: config.RoleIDs.PrdMuted},
		{name: "attachmentmute", id: config.RoleIDs.AttachmentMuted},
		{name: "reactionmute", id: config.RoleIDs.ReactionMuted},
	}

	for _, role := range roles {
		if role.id != 0 {
			addCommand(&Command{
				Name:             role.name,
				Aliases:          []string{strings.Replace(role.name, "mute", "ban", -1)},
				Description:      role.name + " one or more people",
				Usage:            "<Member1 Member2 Member3...> [duration: 1w6h2m] [reason]",
				RequiredArgCount: 1,
				ModOnly:          true,
				OwnerOnly:        false,
				Callback:         makeMuteFunc(role.id),
			})
			unmuteName := strings.Replace(role.name, "mute", "unmute", -1)
			addCommand(&Command{
				Name:             unmuteName,
				Aliases:          []string{strings.Replace(role.name, "mute", "unban", -1)},
				Description:      unmuteName + " one or more people",
				Usage:            "<Member1 Member2 Member3...>",
				RequiredArgCount: 1,
				ModOnly:          true,
				OwnerOnly:        false,
				Callback:         makeUnmuteFunc(role.id, role.name),
			})
		}
	}

	// Reassign mute roles when a user rejoins
	s.AddHandler(func(e *gateway.GuildMemberAddEvent) {
		var mutes []database.Mute

		err := database.DB.Model(&mutes).
			Where("user_id = ?", e.User.ID).
			Where("guild_id = ?", e.GuildID).
			Select()
		if err != nil {
			logger.Println("Failed to retrieve mutes for user ", err)
		} else {
			for _, mute := range mutes {
				if mute.EndDate == -1 || mute.EndDate > time.Now().Unix() {
					err = s.AddRole(e.GuildID, e.User.ID, mute.RoleID, api.AddRoleData{AuditLogReason: api.AuditLogReason(mute.Reason + " - User rejoined")})
					if err != nil {
						logger.Println("Failed to give user back a mute role", err)
					}
				}
			}
		}
	})

	// Start unmute timers
	var mutes []database.Mute

	err := database.DB.Model(&mutes).Where("end_date != ?", -1).Select()
	if err != nil {
		logger.Println("Failed to retrieve mutes", err)
	} else {
		for _, mute := range mutes {
			startUnmuteTimer(mute)
		}
	}
}

func makeMuteFunc(roleID discord.RoleID) func(*CommandContext) (*discord.Message, error) {
	return func(ctx *CommandContext) (*discord.Message, error) {
		cleanedContent := mentionRegex.ReplaceAllString(strings.Join(ctx.Args, " "), "")
		ids := idRegex.FindAllString(cleanedContent, -1)
		cleanedContent = idRegex.ReplaceAllString(cleanedContent, "")

		args := strings.Fields(cleanedContent)
		durationStr := ""
		if len(args) > 0 {
			durationStr = args[0]
		}
		isDuration, duration := parseDuration(durationStr)
		if isDuration {
			args = args[1:]
		}

		reason := ctx.Message.Author.Tag() + ": "
		if len(args) > 0 {
			reason += strings.Join(args, " ")
		} else {
			reason += "No reason specified."
		}
		if isDuration {
			reason += " (For " + durationStr + ")"
		}

		data := api.AddRoleData{
			AuditLogReason: api.AuditLogReason(reason),
		}

		errorCount := 0
		if ids != nil {
			for _, idStr := range ids {
				id, _ := strconv.ParseUint(idStr, 10, 64)
				addMuteRole(ctx.Message.GuildID, discord.UserID(id), roleID, data, &errorCount, isDuration, duration)

			}
		}

		for _, mention := range ctx.Message.Mentions {
			if mention.ID != botUser.ID { // Might be command triggered with mention prefix xD
				addMuteRole(ctx.Message.GuildID, mention.User.ID, roleID, data, &errorCount, isDuration, duration)
			}
		}

		if errorCount == 0 {
			return ctx.Reply("Done!")
		} else {
			return ctx.Reply("I did not manage to give everyone the role. I failed on " + strconv.Itoa(errorCount) + " members :(")
		}
	}
}

func makeUnmuteFunc(roleID discord.RoleID, muteName string) func(*CommandContext) (*discord.Message, error) {
	return func(ctx *CommandContext) (*discord.Message, error) {
		var userID discord.UserID
		if len(ctx.Message.Mentions) > 0 {
			userID = ctx.Message.Mentions[0].ID
		} else {
			match := idRegex.FindString(strings.Join(ctx.Args, " "))
			if match == "" {
				return ctx.Reply("Mention someone!")
			}
			id, _ := strconv.ParseUint(match, 10, 64)
			userID = discord.UserID(id)
		}

		member, err := s.Member(ctx.Message.GuildID, userID)
		if err != nil {
			return ctx.Reply("I couldn't find that Member")
		}
		if !common.HasRole(member.RoleIDs, roleID) {
			return ctx.Reply("That member isnt't " + muteName + "d")
		}

		err = s.RemoveRole(ctx.Message.GuildID, userID, roleID, api.AuditLogReason("Unmuted by"+ctx.Message.Author.Tag()))
		if err != nil {
			return ctx.Reply("Failed to unmute that member")
		}

		_, err = database.DB.Model((*database.Mute)(nil)).
			Where("role_id = ?", roleID).
			Where("user_id = ?", userID).
			Delete()
		logger.LogIfErr(err)

		return ctx.Reply("Done!")
	}
}

func parseDuration(text string) (bool, int64) {
	matches := durationRegex.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return false, 0
	}

	var seconds int64 = 0
	for _, match := range matches {
		i, _ := strconv.ParseInt(match[1], 10, 32)
		switch match[2] {
		case "s", "sec", "secs", "seconds":
			seconds += i
		case "m", "min", "mins", "minutes":
			seconds += i * 60
		case "h", "hours":
			seconds += i * 60 * 60
		case "d", "days":
			seconds += i * 60 * 60 * 24
		case "w", "weeks":
			seconds += i * 60 * 60 * 24 * 7
		default:
			return false, 0
		}
	}

	return true, seconds
}

func addMuteRole(gid discord.GuildID, uid discord.UserID, rid discord.RoleID, data api.AddRoleData, errorCount *int, isDuration bool, duration int64) {
	if err := s.AddRole(gid, uid, rid, data); err != nil {
		*errorCount++
		logger.Println(err)
	} else {
		var endDate int64 = -1
		if isDuration {
			endDate = time.Now().Unix() + duration
		}

		mute := database.Mute{
			UserID:  uid,
			RoleID:  rid,
			GuildID: gid,
			EndDate: endDate,
			Reason:  string(data.AuditLogReason),
		}

		res, err := database.DB.Model(&mute).
			Set("end_date = ?end_date").
			Where("user_id = ?user_id").
			Where("role_id = ?role_id").
			Where("guild_id = ?guild_id").
			Update()

		logger.LogIfErr(err)
		if res.RowsAffected() == 0 { // No entry yet
			_, err := database.DB.Model(&mute).Insert()
			logger.LogIfErr(err)
		}

		if isDuration {
			startUnmuteTimer(mute)
		}
	}
}

func startUnmuteTimer(mute database.Mute) {
	duration := mute.EndDate - time.Now().Unix()
	if duration < 0 {
		duration = 1
	}

	time.AfterFunc(time.Duration(duration*int64(time.Second)), func() {
		err := unmute(mute, "Mute time over")
		if err != nil {
			logger.Println("Failed to unmute user", err)
		}
	})
}

func unmute(mute database.Mute, reason api.AuditLogReason) error {
	err := s.RemoveRole(mute.GuildID, mute.UserID, mute.RoleID, reason)
	if err != nil {
		return err
	} else {
		_, err = database.DB.Model(&mute).
			Where("user_id = ?user_id").
			Where("role_id = ?role_id").
			Delete()
		return err
	}
}
