package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/redis/go-redis/v9"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redisClient := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	key := "counter:request:*"
	mapResult := make(map[string]interface{})

	iter := redisClient.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		get := redisClient.Get(ctx, iter.Val()).Val()
		mapResult[iter.Val()] = get
	}

	keys := make([]string, 0, len(mapResult))
	for k := range mapResult {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("Server : %s\n\n", *serverName))
	for _, v := range keys {
		b.WriteString(fmt.Sprintf("%s : %v\n", v, mapResult[v]))
	}

	_, _ = io.WriteString(w, b.String())
}
