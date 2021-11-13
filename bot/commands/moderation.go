package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"regexp"
	"strconv"
	"strings"
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
				Callback:         makeFunc(role.id),
			})
		}
	}
}

func makeFunc(roleId discord.RoleID) func(*CommandContext) (*discord.Message, error) {
	return func(ctx *CommandContext) (*discord.Message, error) {
		cleanedContent := mentionRegex.ReplaceAllString(strings.Join(ctx.Args, " "), "")
		ids := idRegex.FindAllString(cleanedContent, -1)
		cleanedContent = idRegex.ReplaceAllString(cleanedContent, "")

		args := strings.Fields(cleanedContent)
		durationStr := ""
		if len(args) > 0 {
			durationStr = args[0]
		}
		isDuration, _ := parseDuration(durationStr)
		if isDuration { // TODO: Implement durations
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
				if s.AddRole(ctx.Message.GuildID, discord.UserID(id), roleId, data) != nil {
					errorCount++
				}
			}
		}

		for _, mention := range ctx.Message.Mentions {
			if mention.ID != botUser.ID { // Might be command triggered with mention prefix xD
				if s.AddRole(ctx.Message.GuildID, mention.User.ID, roleId, data) != nil {
					errorCount++
				}
			}
		}

		if errorCount == 0 {
			return ctx.Reply("Done!")
		} else {
			return ctx.Reply("I did not manage to give everyone the role. I failed on " + strconv.Itoa(errorCount) + " members :(")
		}
	}
}

func parseDuration(text string) (bool, int64) {
	matches := durationRegex.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return false, 0
	}

	var seconds int64 = 0
	for _, match := range matches {
		i, _ := strconv.ParseInt(match[0], 10, 32)
		switch match[1] {
		case "s":
		case "sec":
		case "secs":
		case "seconds":
			seconds += i
		case "m":
		case "min":
		case "mins":
		case "minutes":
			seconds += i * 60
		case "h":
		case "hours":
			seconds += i * 60 * 60
		case "d":
		case "days":
			seconds += i * 60 * 60 * 24
		case "w":
		case "weeks":
			seconds += i * 60 * 60 * 24 * 7
		}
	}

	return true, seconds
}
