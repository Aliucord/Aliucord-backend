package commands

import (
	"context"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Aliucord/Aliucord-backend/database"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

var durationRegex = regexp.MustCompile("(\\d+)(s|seconds|secs?|m|mins?|minutes|h|hours|d|days|w|weeks)")

type nameAndID struct {
	name string
	id   discord.RoleID
}

func initModCommands() {
	scamPhrasesOptions := []discord.CommandOptionValue{
		&discord.StringOption{
			OptionName:  "phrase",
			Description: "Phrase",
			Required:    true,
		},
	}
	addCommand(&Command{
		CreateCommandData: api.CreateCommandData{
			Name:        "scamphrases",
			Description: "Add/Remove or list scam phrases",
			Options: []discord.CommandOption{
				&discord.SubcommandOption{
					OptionName:  "list",
					Description: "List scam phrases",
				},
				&discord.SubcommandOption{
					OptionName:  "add",
					Description: "Add scam phrase",
					Options:     scamPhrasesOptions,
				},
				&discord.SubcommandOption{
					OptionName:  "remove",
					Description: "Remove scam phrase",
					Options:     scamPhrasesOptions,
				},
			},
		},
		ModOnly: true,
		Execute: scamPhrasesCommand,
	})

	roles := []nameAndID{
		{name: "supportmute", id: config.RoleIDs.SupportMuted},
		{name: "devmute", id: config.RoleIDs.DevMuted},
		{name: "prdmute", id: config.RoleIDs.PrdMuted},
		{name: "attachmentmute", id: config.RoleIDs.AttachmentMuted},
		{name: "reactionmute", id: config.RoleIDs.ReactionMuted},
	}

	muteOptions := []discord.CommandOption{
		&discord.UserOption{
			OptionName:  "user",
			Description: "user to mute",
		},
		&discord.StringOption{
			OptionName:  "users",
			Description: "users to mute",
		},
		&discord.StringOption{
			OptionName:  "duration",
			Description: "mute duration",
		},
		&discord.StringOption{
			OptionName:  "reason",
			Description: "mute reason",
		},
	}
	unmuteOptions := []discord.CommandOption{
		&discord.UserOption{
			OptionName:  "user",
			Description: "user to unmute",
		},
	}
	for _, role := range roles {
		if role.id != 0 {
			addCommand(&Command{
				CreateCommandData: api.CreateCommandData{
					Name:        role.name,
					Description: role.name + " one or more people",
					Options:     muteOptions,
				},
				ModOnly: true,
				Execute: makeMuteFunc(role.id),
			})
			unmuteName := strings.Replace(role.name, "mute", "unmute", -1)
			addCommand(&Command{
				CreateCommandData: api.CreateCommandData{
					Name:        unmuteName,
					Description: unmuteName,
					Options:     unmuteOptions,
				},
				ModOnly: true,
				Execute: makeUnmuteFunc(role.id, role.name),
			})
		}
	}

	// Reassign mute roles when a user rejoins
	s.AddHandler(func(e *gateway.GuildMemberAddEvent) {
		var mutes []database.Mute

		err := database.DB.NewSelect().
			Model(&mutes).
			Where("user_id = ?", e.User.ID).
			Where("guild_id = ?", e.GuildID).
			Scan(context.Background())
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

	err := database.DB.NewSelect().
		Model(&mutes).
		Where("end_date != ?", -1).
		Scan(context.Background())
	if err != nil {
		logger.Println("Failed to retrieve mutes", err)
	} else {
		for _, mute := range mutes {
			startUnmuteTimer(mute)
		}
	}
}

func scamPhrasesCommand(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
	cmd := d.Options[0]
	subcommand := cmd.Name
	switch subcommand {
	case "list":
		var phrases []database.ScamPhrase
		err := database.DB.NewSelect().
			Model(&phrases).
			Scan(context.Background())
		if err != nil {
			return replyErr(e, "listing scam phrases", err)
		}

		var sb strings.Builder
		for _, phrase := range phrases {
			sb.WriteString(phrase.Phrase)
			sb.WriteRune('\n')
		}

		if sb.Len() == 0 {
			return ephemeralReply(e, "No banned phrases")
		}
		return ephemeralReply(e, "Banned phrases: ```\n"+sb.String()+"```")
	case "add":
		res, err := database.DB.NewInsert().
			Ignore(). // Ignore conflict
			Model(&database.ScamPhrase{Phrase: cmd.Options[0].String()}).
			Exec(context.Background())
		if err != nil {
			return replyErr(e, "adding scam phrase", err)
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			return ephemeralReply(e, "No rows affected. Phrase already added?")
		}

		modules.UpdateScamTitles()
		return reply(e, "Added!")
	case "remove":
		res, err := database.DB.NewDelete().
			Model((*database.ScamPhrase)(nil)).
			Where("phrase = ?", cmd.Options[0].String()).
			Exec(context.Background())
		if err != nil {
			return replyErr(e, "removing scam phrase", err)
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			return ephemeralReply(e, "No rows affected. Phrase not added?")
		}

		modules.UpdateScamTitles()
		return reply(e, "Removed!")
	}
	return reply(e, "No such subcommand: "+subcommand)
}

func makeMuteFunc(roleID discord.RoleID) func(*gateway.InteractionCreateEvent, *discord.CommandInteraction) error {
	return func(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
		userIDs := getUserOrUsersOption(d)
		if len(userIDs) == 0 {
			return ephemeralReply(e, "Provide either `user` or `users` option.")
		}

		durationOption := findOption(d, "duration")
		hasDuration := durationOption != nil
		var duration int64
		var durationStr string
		if hasDuration {
			durationStr = durationOption.String()
			hasDuration, duration = parseDuration(durationStr)
		}

		reason := e.Member.User.Tag() + " - "
		if reasonOption := findOption(d, "reason"); reasonOption == nil {
			reason += "No reason specified."
		} else {
			reason += reasonOption.String()
		}
		if hasDuration {
			reason += " (For " + durationStr + ")"
		}

		data := api.AddRoleData{
			AuditLogReason: api.AuditLogReason(reason),
		}

		errorCount := 0
		mentions := ""
		for _, id := range userIDs {
			if err := addMuteRole(e.GuildID, id, roleID, data, hasDuration, duration); err == nil {
				mentions += " " + id.Mention()
			} else {
				errorCount++
				logger.Println(err)
			}
		}

		if errorCount == 0 {
			return reply(e, "Muted"+mentions+" by "+reason)
		} else {
			return reply(e, "Failed on "+strconv.Itoa(errorCount)+" members, muted"+mentions+" by "+reason)
		}
	}
}

func makeUnmuteFunc(roleID discord.RoleID, muteName string) func(*gateway.InteractionCreateEvent, *discord.CommandInteraction) error {
	return func(e *gateway.InteractionCreateEvent, d *discord.CommandInteraction) error {
		userIDs := common.MapKeys(d.Resolved.Users)
		if len(userIDs) != 1 {
			return ephemeralReply(e, "Mention someone!")
		}

		userID := userIDs[0]
		member, err := s.Member(e.GuildID, userID)
		if err != nil {
			return ephemeralReply(e, "I couldn't find that member")
		}
		if !slices.Contains(member.RoleIDs, roleID) {
			return ephemeralReply(e, "That member isn't "+muteName+"d")
		}

		err = s.RemoveRole(e.GuildID, userID, roleID, api.AuditLogReason("Unmuted by"+e.Member.User.Tag()))
		if err != nil {
			return ephemeralReply(e, "Failed to unmute that member")
		}

		_, err = database.DB.NewDelete().
			Model((*database.Mute)(nil)).
			Where("role_id = ?", roleID).
			Where("user_id = ?", userID).
			Exec(context.Background())
		logger.LogIfErr(err)

		return reply(e, "Unmuted "+member.User.Tag())
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

func addMuteRole(
	gID discord.GuildID,
	uID discord.UserID,
	rID discord.RoleID,
	data api.AddRoleData,
	hasDuration bool,
	duration int64,
) (err error) {
	if err = s.AddRole(gID, uID, rID, data); err == nil {
		var endDate int64 = -1
		if hasDuration {
			endDate = time.Now().Unix() + duration
		}

		mute := database.Mute{
			UserID:  uID,
			RoleID:  rID,
			GuildID: gID,
			EndDate: endDate,
			Reason:  string(data.AuditLogReason),
		}

		res, err := database.DB.NewUpdate().
			Model(&mute).
			Set("end_date = ?end_date").
			Where("user_id = ?user_id").
			Where("role_id = ?role_id").
			Where("guild_id = ?guild_id").
			Exec(context.Background())

		logger.LogIfErr(err)
		if i, _ := res.RowsAffected(); i == 0 { // No entry yet
			_, err = database.DB.NewInsert().Model(&mute).Exec(context.Background())
			logger.LogIfErr(err)
		}

		if hasDuration {
			startUnmuteTimer(mute)
		}
	}
	return
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
	if member, err := s.Member(mute.GuildID, mute.UserID); err == nil && slices.Contains(member.RoleIDs, mute.RoleID) {
		if err = s.RemoveRole(mute.GuildID, mute.UserID, mute.RoleID, reason); err != nil {
			return err
		}
	}

	_, err := database.DB.NewDelete().
		Model(&mute).
		Where("user_id = ?user_id").
		Where("role_id = ?role_id").
		Exec(context.Background())
	return err
}
