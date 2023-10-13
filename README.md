# kong-http-log-to-event
Logging kong http request/response to event streams

```mermaid
sequenceDiagram
    Application->>+Kong Gateway: Call APIs
    Kong Gateway->>+Token Validation Service: Request to Validate and fetch access token
    Token Validation Service->>-Kong Gateway: Response Validation
    Kong Gateway->>+API UpStream: Forward request with extra headers from "Token Validation Service"
    API UpStream->>-Kong Gateway: Get API response
    Kong Gateway->>-Application: Forward API response
    Kong Gateway-->>HTTP Log server: (Async) Submit logs with request and response informations
    HTTP Log server-->>Redis Stream/Kafka/Kinesis: Store logs informations as event message
    Event Message Consumer Service-->>Redis Stream/Kafka/Kinesis: Consume event message
```

## How to test
- Run `make docker-compose-build`
- Run `make docker-compose-up`
- Run `make call-apis`
- Or Run `make load-test` to test with bunch of requests
- Check request log at [`http://localhost:8000/consumer`](http://localhost:8000/consumer)

## Commands

### Check request URL in redis streams
```bash
echo XRANGE mockoon-one - + | redis-cli -h 127.0.0.1 -p 16379 | grep "service" | jq -s '.[].request.url'
```

### Check request headers in redis streams
```bash
echo XRANGE mockoon-one - + | redis-cli -h 127.0.0.1 -p 16379 | grep "service" | jq -s '.[].request.headers'
```

## References
- Original kong plugin, [kong-plugin-http-log-with-body](https://github.com/zenvia/kong-plugin-http-log-with-body)
- Go library for working efficiently with message streams, [Watermill](https://watermill.io)
- Command line tool that runs HTTP requests, [Hurl](https://hurl.dev/)
- HTTP load testing tool and library, [Vegeta](github.com/tsenart/vegeta)


## Documentation

