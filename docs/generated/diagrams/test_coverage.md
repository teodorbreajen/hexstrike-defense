```mermaid
graph TD
    subgraph Test Files
        Tprometheus["prometheus_test.go"]
        Tmetrics["metrics_test.go"]
        Trace["race_test.go"]
        Tsecurity_comprehensive["security_comprehensive_test.go"]
        Tlogger["logger_test.go"]
        Tfuzz["fuzz_test.go"]
        Tproxy_handler["proxy_handler_test.go"]
        Tsecurity["security_test.go"]
        Tcors["cors_test.go"]
        Trate_limiter["rate_limiter_test.go"]
    end
```