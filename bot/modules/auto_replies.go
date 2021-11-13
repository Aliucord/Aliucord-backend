package modules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	modules = append(modules, initAutoReplies)
}

func r(regex string) *regexp.Regexp {
	return regexp.MustCompile(regex)
}

const (
	JustAsk          = "https://dontasktoask.com/"
	CheckThePins     = "<a:checkpins:859804429536198676>"
	MentionHelp      = "Rule 9: Don't dm or mention for support"
	ElaborateHelp    = "We can't help you if you don't tell us your issue. "
	PluginDownloader = "This is already a coreplugin of Aliucord. Update Aliucord if it's missing."
	FreeNitro        = "Not possible. Nitrospoof exists for \"free\" emotes, for anything else buy nitro."
	Usage            = "Go to the plugin's repository and read the readme. Chances are the dev added a description."
)

func initAutoReplies() {
	cfg := config.AutoReplyConfig
	if !cfg.Enabled {
		return
	}

	var (
		PRD        = fmt.Sprintf("%s 👉 <#%s>", CheckThePins, cfg.PRD)
		FindPlugin = fmt.Sprintf("Look in <#%s> and <#%s>. If it doesn't exist, then %s in <#%s>", cfg.PluginsList, cfg.NewPlugins, CheckThePins, cfg.PRD)
	)

	var autoRepliesString = map[string]string{
		"free nitro":       FreeNitro,
		"animated profile": FreeNitro,
		"animated avatar":  FreeNitro,
		"nitro perks":      FreeNitro,
		"how do i use":     Usage,
		"a plugin to":      PRD,
		"can you make":     PRD,
	}

	var autoRepliesRegex = map[*regexp.Regexp]string{
		r("(?i)^help$"):                          ElaborateHelp,
		r("(?i)<@!?\\d{2,19}> help"):             MentionHelp,
		r("(?i)help <@!?\\d{2,19}>"):             MentionHelp,
		r("(?i)give me .+ link"):                 FindPlugin,
		r("(?i)link to .+ plugin"):               FindPlugin,
		r("(?i)where is .+ plugin"):              FindPlugin,
		r("(?i)is there(?: a )? .+ plugin"):      FindPlugin,
		r("(?i)can (?:anyone|you) help(?: me)?"): JustAsk,
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		for _, role := range msg.Member.RoleIDs {
			if common.HasRole(cfg.IgnoredRoles, role) {
				return
			}
		}

		for regex, reply := range autoRepliesRegex {
			if regex.MatchString(msg.Content) {
				_, err := s.SendTextReply(msg.ChannelID, reply, msg.ID)
				logger.LogIfErr(err)
				return
			}
		}

		content := strings.ToLower(msg.Content)
		for trigger, reply := range autoRepliesString {
			if !strings.Contains(content, trigger) {
				return
			}

			_, err := s.SendTextReply(msg.ChannelID, reply, msg.ID)
			logger.LogIfErr(err)

			return
		}
	})
}
