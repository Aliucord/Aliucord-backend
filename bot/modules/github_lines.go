package modules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/valyala/fasthttp"
)

const maxContentLength = 5e6 // 5MB
const maxCharCount = 1000

var client *fasthttp.Client
var githubLinesRegex = regexp.MustCompile(`https?://github\.com/([A-Za-z0-9\-_.]+)/([A-Za-z0-9\-_.]+)/(?:blob|tree)/(\S+?)/(\S+?)(\.\S+)?#L(\d+)[-~]?L?(\d*)`)

func init() {
	modules = append(modules, initGithubLines)
	client = &fasthttp.Client{}
}

func initGithubLines() {
	s.AddHandler(func(msg *gateway.MessageCreateEvent) {
		matches := githubLinesRegex.FindAllStringSubmatch(msg.Content, -1)
		if len(matches) != 0 {
			var sb strings.Builder
			for _, match := range matches {
				author, repo, branch, path, ext, start, end := match[1], match[2], match[3], match[4], match[5], match[6], match[7]
				rawUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s%s", author, repo, branch, path, ext)

				if shouldSkip(rawUrl) {
					continue
				}

				content := getContent(rawUrl, start, end)
				if content == "" {
					continue
				}

				fileNameIdx := strings.LastIndex(path, "/") + 1
				fileName := path
				if fileNameIdx != 0 {
					fileName = path[fileNameIdx:]
				}
				sb.WriteString("**" + fileName + ext + "#" + start)
				if end != "" {
					sb.WriteString("-" + end)
				}
				sb.WriteString("**\n")

				sb.WriteString("```")
				if len(ext) > 1 {
					lang := ext[1:]
					if common.IsAlpha(lang) {
						sb.WriteString(lang)
					}
				}

				sb.WriteString("\n" + content + "\n```\n")
			}

			length := sb.Len()
			if length > 0 && length < maxCharCount {
				_, err := s.SendMessageComplex(msg.ChannelID, api.SendMessageData{
					Content:         sb.String(),
					AllowedMentions: &api.AllowedMentions{RepliedUser: option.False},
					Reference:       &discord.MessageReference{MessageID: msg.ID},
				})
				logger.LogWithCtxIfErr("embedding github lines", err)
			}
		}

	})
}

// Check if url is invalid or file larger than maxContentLength
func shouldSkip(url string) bool {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodHead)
	err := client.Do(req, res)
	if err != nil {
		return true
	}

	return res.Header.ContentLength() > maxContentLength
}

// Fetch content from url and trim according to startStr and endStr
// Empty string is returned if failed to fetch, startStr is empty or startStr or endStr are not valid integers / out of bounds
func getContent(url, startStr, endStr string) string {
	start, err := strconv.Atoi(startStr)
	if err != nil || start < 1 {
		return ""
	}
	end := 0
	if endStr != "" {
		end, err = strconv.Atoi(endStr)
		if err != nil || end < 0 {
			return ""
		}
		if start > end {
			start, end = end, start
		} else if start == end {
			end = 0
		}
	}

	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	err = client.Do(req, res)
	if err != nil || res.StatusCode() != 200 {
		return ""
	}

	content := string(res.Body())
	lines := strings.Split(content, "\n")
	lineCount := len(lines)

	if end == 0 && lineCount >= start {
		return lines[start-1]
	}
	if lineCount >= end {
		return strings.Join(lines[start-1:end], "\n")
	}

	return ""
}
