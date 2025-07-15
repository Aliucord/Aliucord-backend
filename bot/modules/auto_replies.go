package modules

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func init() {
	modules = append(modules, initAutoReplies)
}

func r(regex string) *regexp.Regexp {
	return regexp.MustCompile("(?i)" + regex)
}

var lastReplyCache = map[discord.UserID]string{}

func reply(msg *gateway.MessageCreateEvent, reply string) {
	if lastReplyCache[msg.Author.ID] == reply {
		return
	}
	_, err := s.SendTextReply(msg.ChannelID, reply, msg.ID)
	logger.LogIfErr(err)
	lastReplyCache[msg.Author.ID] = reply
}

const (
	//JustAsk        = "https://dontasktoask.com/"
	CheckThePins     = "<a:checkpins:859804429536198676>"
	MentionHelp      = "Rule 9: Don't dm or mention for support"
	//ElaborateHelp  = "We can't help you if you don't tell us your issue. "
	InstallPlugins   = "https://aliucord.com/files/tut/InstallPlugins.mp4" // people would rather watch a video than opening the docu so
	InstallThemes    = "https://aliucord.com/files/tut/InstallThemes.mp4"  // ven owes me a million dollars, now with themer install
	CreateThemes     = "Read this documentation: https://github.com/Aliucord/documentation/tree/main/theme-dev"
	ThemeSounds      = "https://aliucord.com/files/tut/ThemeSounds.mp4" // sounds
	// GetSound      = "https://cdn.discordapp.com/attachments/875213883776847873/1007312170868019210/20220811_103613.mp4"
	FullTransparency = "1. Are you using a theme that requires full transparency? If the answer is no, then that's the problem. Normally the description says what transparency you need to use.\n2. Are you using a custom ROM? If the answer is yes, then we can't do anything about it."
	ThemerHusk       = `Make sure you are using correct transparency settings for your theme and <a:checkpins:859804429536198676> in <#875213883776847873> for fixed versions of themes.
If that doesn't help, Themer might just be broken for you, which happens often on new Android versions and custom ROMs`
	SearchThemes     = "Check <#824357609778708580> and search on there, maybe there's the theme you want"
	AliuCrash        = `Make sure to install Aliucord from the lastest Manager and check for existing solutions in <#857624431148138518> as well as <#811257821378904064>.
If this doesn't solve your problem send crashlogs (check Crashes in Settings, and copy the most recent), if there aren't any crashlogs then you can try to send a logcat (check <https://pastebin.com/pNhXwhrd>)`
	FreeNitro        = `Not possible, but there are plugins that mimic nitro functionality:
1. **NitroSpoof** for free nitro emojis as images.
2. **FakeStickers** for free nitro stickers as images.
3. **UserPFP** for free animated profile picture visible to other people using it. In order to use it, follow [this guide](<https://pastebin.com/wbHcbRdt>).
4. **UserBG** for free animated/custom banner visible to other people using it. In order to use it, follow [this guide](<https://pastebin.com/BF8D7dtv>).
5. **CustomBadges** for profile badges visible to you. In order to use it, check this [list of drawables](<https://gist.github.com/Vendicated/65775d2868eb7a1e05a65c2a8d5784fc>).
**To install them, just hold this message (NOT THE LINKS) and u will have a option to do so.**
[1](<https://github.com/X1nto/AliucordPlugins/blob/builds/NitroSpoof.zip?raw=true>) [2](<https://github.com/RhythmLunatic/aliucord-plugins/blob/builds/FakeStickers.zip?raw=true>) [3](<https://github.com/OmegaSunkey/awesomeplugins/blob/builds/UserBG.zip?raw=true>) [4](<https://github.com/wingio/plugins/blob/builds/CustomBadges.zip?raw=true>)`
	Usage            = "Read the plugin's description in <#811275162715553823> or <#845784407846813696>. You can also go to the plugin's repository and look at information in the readme."
	BetterInternet   = "This happens when you have an old/misbehaving router. Use mobile data (~120mb usage) or maybe a VPN (*or just get better internet*)."
	PluginDownloader = "PluginDownloader is now a part of Aliucord. (It won't be present in the plugin list). If the option to download plugins is still missing, reinstall Aliucord."
	WhereAliucord    = "https://github.com/Aliucord/Manager#Installation"
	Backports        = `Aliucord uses (and always will use) old version of Discord (126.21), which means some things will work differently that in current Discord.
For a list of new Discord features that have been backported to Aliucord see <#858409546791518237>.
If you prefer the recent look see this [list of other Discord mods ](<https://github.com/Discord-Client-Encyclopedia-Management/Discord3rdparties>)`
	SlashCommands    = "Aliucord currently lacks full bot slash command support, but you can install an experimental slash command fix by following the instructions in <#1210929772683599882> pins."
)

func initAutoReplies() {
	cfg := config.AutoReplyConfig
	if !cfg.Enabled {
		return
	}

	PRD := fmt.Sprintf("%s 👉 <#%s>", CheckThePins, cfg.PRD)
	FindPlugin := fmt.Sprintf("Search in <#%s> and <#%s> for keywords related to the plugin. If you are sure it doesn't exist, then you should %s in <#%s>",
		cfg.PluginsList, cfg.NewPlugins, CheckThePins, cfg.PRD)

	autoRepliesString := map[string]string{
		//"can you make":          PRD,
		"how do i use":          Usage,
		"free nitro":            FreeNitro,
		"handshake exception":   BetterInternet,
		"connection terminated": BetterInternet,
	}

	autoRepliesRegex := map[*regexp.Regexp]string{
		//r("^(?:i need )?help(?: me)?$"):                                                                         ElaborateHelp,
		r("help <@!?\\d{17,19}>|<@!?\\d{17,19}> help"):                                                            MentionHelp,
		r("animated (profile|avatar|pfp)"):                                                                        FreeNitro,
		r("is there a plugin"):                                                                                    FindPlugin,
		r("^where(?: i)?'?s(?: the )?.+ plugin$"):                                                                 FindPlugin,
		//r("^can (?:someone|anybody|anyone|you) help(?: me)?\\??$"):                                              JustAsk,
		r("(can.?not|can'?t) (download|find|get) plugin downloader"):                                              PluginDownloader,
		r("where( i)?'?s( the)? plugin downloader"):                                                               PluginDownloader,
		r("(?:where|how) (?:to|do i|do you) (?:install|download|get) a? ?plugins?"):                               InstallPlugins,
		r("how (?:to|do i|do you|i) (?:install|download|apply|get|use) a? ?themes?"):                              InstallThemes,
		r("how (?:to|do i|do you|can i) (?:create|make|do) (?:a |my )?(?:own |custom )?themes?"):                  CreateThemes,
		r("how (?:to|do i|do you|can i|put) (?:change|upload|add|set) (?:sounds?|custom sounds?)"):                ThemeSounds,
		// r("how (?:to|do i|do you|can i) get sounds?(?: url| link)?"):                                           GetSound,
		r("(?:does anyone know|is there) an? (?:\\w.+)?theme"):                                                    SearchThemes,
		r("aliucord (?:\\w+ )?(?:is |keeps? )?(?:crash|stop)"):                                                    AliuCrash,
		r("full transparency (?:is?.?not|isn\\'?t|will not|doesn\\'t|does not) work"):                             FullTransparency,
		r("theme (?:(?:is?.?not|isn\\'?t|will not|doesn\\'?t|does not) work)|(?:(?: is)? broken)"):                ThemerHusk,
		r("where(?: \\w+){0,7} aliucord"):                                                                         WhereAliucord,
		r("why (?:is|are)(?: [\\w,.'-]+){0,5} (?:not like|different from|different than)(?: in)? discord"):        Backports,
		r("why (?:is|are)(?: [\\w,.'-]+){0,5} (?:broken|not work(?:ing)?|different) in aliucord"):                 Backports,
		r("why (?:is|are|isn'?t|aren'?t|don'?t|doesn'?t|can(?:no|'?t))(?: [\\w,.'-]+){1,7} in (?:aliu|dis)cord"):  Backports,
		r("why(?: is| does)? aliucord (?:looking|look|looks)(?: [\\w,.'-]+){0,4} old"):                            Backports,
		r("(?:slash|bot|app)(?: slash| bot| app) commands (?:don'?t work|not work|broken)"):                       SlashCommands,
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if msg.Member == nil || len(msg.Attachments) > 0 || (msg.ReferencedMessage != nil && msg.ReferencedMessage.Author.ID == msg.Author.ID) || msg.Author.Bot || strings.HasPrefix(msg.Content, "Quick Aliucord ") {
			return
		}

		c, err := s.Channel(msg.ChannelID)
		if err == nil {
			if c.ID != cfg.PRD && c.ParentID != cfg.SupportCategory {
				return
			}
		} else {
			logger.Println(err)
		}

		for _, role := range msg.Member.RoleIDs {
			if slices.Contains(cfg.IgnoredRoles, role) {
				return
			}
		}

		for regex, value := range autoRepliesRegex {
			if regex.MatchString(msg.Content) {
				reply(msg, value)
				return
			}
		}

		content := strings.ToLower(msg.Content)
		for trigger, value := range autoRepliesString {
			if strings.Contains(content, trigger) {
				reply(msg, value)
				return
			}
		}
	})
}
