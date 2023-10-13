package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	stdHttp "net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"

	wMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-http/pkg/http"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/spf13/pflag"
)

var (
	redisAddr       = pflag.String("redis", "redis.my.internal:6379", "The address of the redis stream")
	httpAddr        = pflag.String("http", ":80", "The address for the http subscriber")
	endpointsString = pflag.StringToInt64("endpoints", map[string]int64{"http-log": 100}, "The endpoints to capture data")
)

func main() {
	pflag.Parse()
	if _, ok := (*endpointsString)["http-log"]; !ok {
		(*endpointsString)["http-log"] = 100
	}

	logger := watermill.NewStdLogger(true, true)

	httpRouter := chi.NewRouter()
	httpRouter.Use(
		// Middleware basic auth
		func(next stdHttp.Handler) stdHttp.Handler {
			fn := func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
				realm := "Http Log"
				creds := map[string]string{
					"my_user": "my_password",
				}

				user, pass, ok := r.BasicAuth()
				basicAuthFailed := func(w stdHttp.ResponseWriter, realm string) {
					w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
					w.WriteHeader(stdHttp.StatusUnauthorized)
				}

				if !ok {
					basicAuthFailed(w, realm)
					return
				}

				credPass, credUserOk := creds[user]
				if !credUserOk || subtle.ConstantTimeCompare([]byte(pass), []byte(credPass)) != 1 {
					basicAuthFailed(w, realm)
					return
				}

				next.ServeHTTP(w, r)
			}

			return stdHttp.HandlerFunc(fn)
		},
		// Middleware to simulate delay
		func(next stdHttp.Handler) stdHttp.Handler {
			fn := func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
				delay := r.URL.Query().Get("delay")
				intDelay, e := strconv.Atoi(delay)
				if e == nil {
					ctx := r.Context()
					delayTime := time.Duration(intDelay) * time.Second
					select {
					case <-ctx.Done():
						return

					case <-time.After(delayTime):
					}
				}

				next.ServeHTTP(w, r)
			}

			return stdHttp.HandlerFunc(fn)
		},
	)

	subscriber, err := http.NewSubscriber(
		*httpAddr,
		http.SubscriberConfig{
			Router: httpRouter,
			UnmarshalMessageFunc: func(topic string, request *stdHttp.Request) (*message.Message, error) {
				b, err := io.ReadAll(request.Body)
				if err != nil {
					return nil, errors.Wrap(err, "cannot read body")
				}

				return message.NewMessage(watermill.NewUUID(), b), nil
			},
		},
		logger,
	)

	if err != nil {
		panic(err)
	}

	pubClient := redis.NewClient(&redis.Options{
		Addr: *redisAddr,
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:  pubClient,
			Maxlens: *endpointsString,
		},
		logger,
	)

	if err != nil {
		panic(err)
	}

	r, err := message.NewRouter(
		message.RouterConfig{},
		logger,
	)

	if err != nil {
		panic(err)
	}

	r.AddMiddleware(
		wMiddleware.Recoverer,
		wMiddleware.CorrelationID,
	)

	r.AddPlugin(plugin.SignalsHandler)

	for k := range *endpointsString {
		r.AddHandler(
			fmt.Sprintf("%s-%s", "http_to_redis", k),
			fmt.Sprintf("/%s", k),
			subscriber,
			k,
			publisher,
			func(msg *message.Message) ([]*message.Message, error) {
				var js json.RawMessage

				if err := json.Unmarshal(msg.Payload, &js); err != nil {
					r.Logger().Error(
						"Handler Error",
						errors.Wrap(err, "cannot unmarshal message"),
						nil,
					)

					return nil, nil
				}

				if js == nil {
					r.Logger().Error(
						"Handler Error",
						errors.New("empty object kind"),
						nil,
					)
					return nil, nil
				}

				msg.Metadata.Set("name", "HTTPLog")

				return []*message.Message{msg}, nil
			},
		)
	}

	go func() {
		// HTTP server needs to be started after router is ready.
		<-r.Running()
		_ = subscriber.StartHTTPServer()
	}()

	_ = r.Run(context.Background())
}
