package modules

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/gateway"
	"golang.org/x/exp/slices"
)

func init() {
	modules = append(modules, initAntiZip)
}

var blacklistedExts = []string{"zip", "exe", "dll", "jar", "deb", "msi", "tar", "gz", "gzip", "bz2", "xz", "apk", "apks", "xapk", "cab", "iso", "img", "rpm", "dmg", "com", "bat", "sh", "zst"}

func initAntiZip() {
	if !config.AntiZip.Enabled {
		return
	}

	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		if len(msg.Attachments) == 0 || msg.Member == nil {
			return
		}

		if !slices.Contains(config.AntiZip.TargetChannels, msg.ChannelID) {
			return
		}

		for _, role := range msg.Member.RoleIDs {
			if slices.Contains(config.RoleIDs.IgnoredRoles, role) {
				return
			}
		}

		for _, attachment := range msg.Attachments {
			for _, ext := range blacklistedExts {
				if strings.HasSuffix(attachment.Filename, ext) {
					err := s.DeleteMessage(msg.ChannelID, msg.ID, "Sent disallowed attachment type")
					if err != nil {
						logger.Println(err)
					}
					return
				}
			}
		}
	})
}
