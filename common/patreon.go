package common

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	PatreonSecret []byte
)

// Generate Type from json go brrrrr
// See https://docs.patreon.com/#webhooks

type patreonPayload struct {
	Data struct {
		Attributes struct {
			AmountCents    int         `json:"amount_cents"`
			CreatedAt      time.Time   `json:"created_at"`
			DeclinedSince  interface{} `json:"declined_since"`
			PatronPaysFees bool        `json:"patron_pays_fees"`
			PledgeCapCents interface{} `json:"pledge_cap_cents"`
		} `json:"attributes"`
		Id            string `json:"id"`
		Relationships struct {
			Address struct {
				Data interface{} `json:"data"`
			} `json:"address"`
			Card struct {
				Data interface{} `json:"data"`
			} `json:"card"`
			Creator struct {
				Data struct {
					Id   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"creator"`
			Patron struct {
				Data struct {
					Id   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"patron"`
			Reward struct {
				Data struct {
					Id   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"reward"`
		} `json:"relationships"`
		Type string `json:"type"`
	} `json:"data"`
	Included []interface{} `json:"included"`
}

func ValidateSignature(message []byte, expectedMessage string) bool {
	mac := hmac.New(md5.New, PatreonSecret)
	mac.Write(message)
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum) == expectedMessage
}

func HandlePatreonWebhook(ctx *fasthttp.RequestCtx) {
	signature := string(ctx.Request.Header.Peek("X-Patreon-Signature"))
	if signature == "" {
		FailRequest(ctx, fasthttp.StatusBadRequest)
		return
	}
	if !ValidateSignature(ctx.PostBody(), signature) {
		FailRequest(ctx, fasthttp.StatusUnauthorized)
		return
	}

	var payload *patreonPayload
	err := json.Unmarshal(ctx.PostBody(), &payload)
	if err != nil {
		FailRequest(ctx, fasthttp.StatusBadRequest)
	}

	switch string(ctx.Request.Header.Peek("X-Patreon-Event")) {
	// TODO these events are actually deprecated, patreon doc moment, switch to ig members:pledge:create, members:pledge:update and members:pledge:delete
	case "pledges:create":
		println("Pledge Create")
	case "pledges:update":
		println("Pledge Update")
	case "pledges:delete":
		println("Pledge Delete")
	default:
		FailRequest(ctx, fasthttp.StatusBadRequest)
	}
}
