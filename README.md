# ipfs-telemetry

## config

```json
{
    "Telemetry": {
        // default: false
        "Enabled": true
        
        // default: "public"
        // possible values: public,restricted,disabled
        "AccessType": "public"

        // only used with access type restricted
        // array of peer ids allowed to collect telemetry
        // default: []
        "Whitelist": [], 

        // how often metrics are collected
        // default: "20s"
        "MetricsPeriod": "5s",

        // how long metrics/events get stored for
        // default: "30m"
        "WindowDuration": "60m",
    }
}
```
