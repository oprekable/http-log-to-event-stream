package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"

	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"golang.org/x/sync/errgroup"

	"github.com/spf13/pflag"
)

var (
	serviceName     = pflag.String("service", "consumer", "The service name")
	serverName      = pflag.String("server", "server-0", "The server name")
	redisAddr       = pflag.String("redis", "redis.my.internal:6379", "The address of the redis stream")
	httpAddr        = pflag.String("http", ":80", "The address for the http subscriber")
	endpointsString = pflag.StringToInt64("endpoints", map[string]int64{"http-log": 100}, "The endpoints to capture data")
)

const keyServerAddr = "serverAddr"
const keyServiceName = "serviceName"

type HTTPHandlerMap map[string]func(http.ResponseWriter, *http.Request)

func runHTTPServer(ctx context.Context, eg *errgroup.Group, httpAddr string, httpHandlerMap HTTPHandlerMap, serviceName string) {
	mux := http.NewServeMux()

	for k, v := range httpHandlerMap {
		mux.HandleFunc(k, v)
	}

	serverHttp := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			ctx = context.WithValue(ctx, keyServiceName, serviceName)
			return ctx
		},
	}

	eg.Go(func() error {
		err := serverHttp.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("[start] HTTP server closed\n")
		} else if err != nil {
			fmt.Printf("[start] HTTP server error listening for server: %s\n", err)
		}

		return err
	})

	eg.Go(func() (err error) {
		select {
		case <-ctx.Done():
			if err = serverHttp.Shutdown(context.Background()); err != nil {
				fmt.Printf("[shutdown] HTTP server failed to shutting down")
			}

			fmt.Printf("[shutdown] HTTP server")
			return
		}
	})
}

func runConsumerServer(ctx context.Context, eg *errgroup.Group, redisAddr string, endpointsString map[string]int64) {
	logger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)

	if err != nil {
		panic(err)
	}

	pubSubClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:  pubSubClient,
			Maxlens: endpointsString,
		},
		logger,
	)

	if err != nil {
		panic(err)
	}

	router.AddMiddleware(middleware.Recoverer)

	for k := range endpointsString {
		_, err = cqrs.NewFacade(cqrs.FacadeConfig{
			GenerateEventsTopic: func(eventName string) string {
				return k
			},
			EventsPublisher: publisher,
			EventHandlers: func(commandBus *cqrs.CommandBus, eventBus *cqrs.EventBus) []cqrs.EventHandler {
				return []cqrs.EventHandler{
					NewGetRequestURLHandler(
						pubSubClient,
						k,
					),
					NewGetResponseHTTPCodeHandler(
						pubSubClient,
						k,
						http.StatusOK,
					),
					NewGetResponseHTTPCodeHandler(
						pubSubClient,
						k,
						http.StatusInternalServerError,
					),
				}
			},
			EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
				h := strings.SplitN(handlerName, "_", 2)
				return redisstream.NewSubscriber(
					redisstream.SubscriberConfig{
						Client:        pubSubClient,
						ConsumerGroup: fmt.Sprintf("%s_%s_%s", *serviceName, h[0], k),
					},
					logger,
				)
			},
			Router: router,
			CommandEventMarshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: logger,
		})

		if err != nil {
			panic(err)
		}
	}

	eg.Go(func() error {
		err = router.Run(ctx)
		if err != nil {
			panic(err)
		}

		return err
	})
}

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	sigTrap := TermSignalTrap()

	defer func() {
		cancelCtx()
	}()

	pflag.Parse()
	if _, ok := (*endpointsString)["http-log"]; !ok {
		(*endpointsString)["http-log"] = 100
	}

	httpHandlerMap := HTTPHandlerMap{
		"/": rootHandler,
	}

	runHTTPServer(ctx, eg, *httpAddr, httpHandlerMap, *serviceName)
	runConsumerServer(ctx, eg, *redisAddr, *endpointsString)

	eg.Go(func() error {
		return sigTrap.Wait(ctx)
	})

	if err := eg.Wait(); err != nil && err != ErrTermSig {
		panic("graceful shutdown successfully finished")
	}
}
