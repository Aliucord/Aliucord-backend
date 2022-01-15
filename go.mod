module github.com/Aliucord/Aliucord-backend

go 1.16

replace github.com/diamondburned/arikawa/v3 => github.com/Juby210/arikawa/v3 v3.0.0-20220115181056-b929dc424a60

require (
	github.com/Juby210/admh v0.0.4
	github.com/Juby210/gplayapi-go v0.0.4
	github.com/diamondburned/arikawa/v3 v3.0.0-rc.4
	github.com/mattn/anko v0.1.9-0.20200521053103-0d30f07e8629
	github.com/uptrace/bun v1.0.21
	github.com/uptrace/bun/dialect/pgdialect v1.0.21
	github.com/uptrace/bun/driver/pgdriver v1.0.21
	github.com/valyala/fasthttp v1.32.0
)
