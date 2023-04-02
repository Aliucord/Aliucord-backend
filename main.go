package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Aliucord/Aliucord-backend/bot"
	"github.com/Aliucord/Aliucord-backend/bot/modules"
	"github.com/Aliucord/Aliucord-backend/common"
	"github.com/Aliucord/Aliucord-backend/database"
	"github.com/Aliucord/Aliucord-backend/updateTracker"
	"github.com/valyala/fasthttp"
)

type fastHttpLogger struct{}

func (*fastHttpLogger) Printf(string, ...interface{}) {}

func main() {
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	var config *common.Config
	err = json.NewDecoder(f).Decode(&config)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	database.InitDB(config.Database)
	if config.Bot.Enabled {
		modules.UpdateScamTitles()
		bot.StartBot(config)
		defer bot.StopBot()
	}
	updateTracker.StartUpdateTracker(config)

	log.Println("Starting http server at port", config.Port)
	fs := &fasthttp.FS{
		Root:        config.ApkCacheDir,
		PathRewrite: fasthttp.NewPathSlashesStripper(2),
	}
	apkFsHandler := fs.NewRequestHandler()
	staticFs := &fasthttp.FS{Root: "static", IndexNames: []string{"index.html"}}
	staticFsHandler := staticFs.NewRequestHandler()
	server := fasthttp.Server{
		Logger: &fastHttpLogger{},
		Handler: func(ctx *fasthttp.RequestCtx) {
			path := string(ctx.Path())
			switch path {
			case "/links/github":
				ctx.Redirect("https://github.com/Aliucord", fasthttp.StatusMovedPermanently)
			case "/links/discord":
				ctx.Redirect("https://discord.gg/EsNDvBaHVU", fasthttp.StatusMovedPermanently)
			case "/download/discord":
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
				url, err := updateTracker.GetDownloadURL(version, string(ctx.QueryArgs().Peek("split")), false)
				if err != nil {
					ctx.SetStatusCode(fasthttp.StatusNotFound)
					ctx.WriteString("apk not found: " + err.Error())
					return
				}
				ctx.Redirect(url, fasthttp.StatusFound)
			default:
				if strings.HasPrefix(path, "/download/direct/") {
					apkFsHandler(ctx)
				} else {
					staticFsHandler(ctx)
				}
			}
		},
	}
	err = server.ListenAndServe(":" + config.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func missingParams(ctx *fasthttp.RequestCtx, params []string) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	ctx.WriteString("missing required query params:\n   " + strings.Join(params, "\n   "))
}
