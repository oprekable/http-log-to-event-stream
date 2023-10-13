package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/buger/jsonparser"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type HTTPLog struct {
	json.RawMessage
}

type getRequestURLHandler struct {
	redisClient *redis.Client
	suffix      string
}

func NewGetRequestURLHandler(
	redisClient *redis.Client,
	suffix string,
) cqrs.EventHandler {
	returnData := &getRequestURLHandler{
		redisClient: redisClient,
		suffix:      suffix,
	}

	return returnData
}

func (h getRequestURLHandler) HandlerName() string {
	return fmt.Sprintf("%s_%s", "getRequestURLHandler", h.suffix)
}

func (h getRequestURLHandler) NewEvent() interface{} {
	return &HTTPLog{}
}

func (h getRequestURLHandler) Handle(ctx context.Context, event interface{}) error {
	if message.SubscribeTopicFromCtx(ctx) == h.suffix {
		e := event.(*HTTPLog)
		jString, _ := json.Marshal(&e)
		getRequestURL, _ := jsonparser.GetString(jString, "request", "url")
		getRequestMethod, _ := jsonparser.GetString(jString, "request", "method")
		fmt.Printf("\n%s : %s", h.HandlerName(), getRequestURL)

		parse, _ := url.Parse(getRequestURL)

		h.redisClient.Incr(
			ctx,
			fmt.Sprintf("counter:request:endpoint:%s:%s", getRequestMethod, parse.Path),
		)
	}
	return nil
}

type getResponseHTTPCodeHandler struct {
	redisClient *redis.Client
	suffix      string
	httpCode    int
}

func NewGetResponseHTTPCodeHandler(
	redisClient *redis.Client,
	suffix string,
	httpCode int,
) cqrs.EventHandler {
	returnData := &getResponseHTTPCodeHandler{
		redisClient: redisClient,
		suffix:      suffix,
		httpCode:    httpCode,
	}

	return returnData
}

func (h getResponseHTTPCodeHandler) HandlerName() string {
	return fmt.Sprintf("%s_%s_%d", "getResponseHTTPCodeHandler", h.suffix, h.httpCode)
}

func (h getResponseHTTPCodeHandler) NewEvent() interface{} {
	return &HTTPLog{}
}

func (h getResponseHTTPCodeHandler) Handle(ctx context.Context, event interface{}) error {
	if message.SubscribeTopicFromCtx(ctx) == h.suffix {
		e := event.(*HTTPLog)
		jString, _ := json.Marshal(&e)
		getRequestURL, _ := jsonparser.GetString(jString, "request", "url")
		getResponseCode, _ := jsonparser.GetInt(jString, "response", "status")
		symbol := "✖"
		if int(getResponseCode) == h.httpCode {
			symbol = "✔"
		}
		fmt.Printf("\n%s : %s - %d %s", h.HandlerName(), getRequestURL, getResponseCode, symbol)
		h.redisClient.Incr(
			ctx,
			fmt.Sprintf("counter:request:response_code:%d", getResponseCode),
		)
	}
	return nil
}
