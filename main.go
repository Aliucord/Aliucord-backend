package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Aliucord/Aliucord-backend/bot"
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Aliucord/Aliucord-backend/database"
	"github.com/Aliucord/Aliucord-backend/updateTracker"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var redirects = map[string]string{
	"github":  "https://github.com/Aliucord",
	"discord": "https://discord.gg/EsNDvBaHVU",
	"patreon": "https://patreon.com/Aliucord",
}

type fastHttpLogger struct{}

func (*fastHttpLogger) Printf(string, ...interface{}) {}

func main() {
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var config *common.Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	common.PatreonSecret = []byte(config.PatreonSecret)

	database.InitDB(config.Database)
	if config.Bot.Enabled {
		bot.StartBot(config)
		defer bot.StopBot()
	}
	updateTracker.StartUpdateTracker(config)

	log.Println("Starting http server at port", config.Port)
	fs := &fasthttp.FS{
		Root:        config.UpdateTracker.DiscordJADX.WorkDir + "/apk",
		PathRewrite: fasthttp.NewPathSlashesStripper(2),
	}
	fsHandler := fs.NewRequestHandler()

	r := router.New()

	r.GET("/", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(
			ctx,
			"<html><head><title>Aliucord</title></head><body>High quality™ temp page.<br><a href=\"/links/github\">[GitHub Org]</a> <a href=\"/links/discord\">[Discord Server]</a></body></html>",
		)
	})

	r.GET("/links/{platform}", func(ctx *fasthttp.RequestCtx) {
		redirect := redirects["platform"]
		if redirect != "" {
			ctx.Redirect(redirect, fasthttp.StatusMovedPermanently)
		} else {
			common.FailRequest(ctx, fasthttp.StatusNotFound)
		}
	})

	r.GET("/download/discord", func(ctx *fasthttp.RequestCtx) {
		v := string(ctx.QueryArgs().Peek("v"))
		if v == "" {
			missingParams(ctx, []string{"v - discord version code"})
			return
		}
		version, err := strconv.Atoi(v)
		if err != nil {
			missingParams(ctx, []string{"v - discord version code"})
			return
		}
		url, err := updateTracker.GetDownloadURL(version, false)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.WriteString("apk not found: " + err.Error())
			return
		}
		ctx.Redirect(url, fasthttp.StatusFound)
	})

	r.GET("/download/direct/{filepath:*}", fsHandler)
	r.POST("/patreon-webhook", common.HandlePatreonWebhook)

	server := fasthttp.Server{
		Logger:  &fastHttpLogger{},
		Handler: r.Handler,
	}
	log.Fatal(server.ListenAndServe(":" + config.Port))
}

func missingParams(ctx *fasthttp.RequestCtx, params []string) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	ctx.WriteString("missing required query params:\n   " + strings.Join(params, "\n   "))
}
