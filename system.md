# system

Subject: `discovery`
```json5
{
    "id": "<peerid>",   // Peer ID
    "addresses": [      // Array of multi addresses
        "<multiaddr>"
    ]
}
```

Subject: `telemetry.raw`
```json5
{
    "object_key": "<s3 file name>"
}
```

S3 naming scheme: `<year>/<month>/<day>/<hour>/<unix-timestamp>-<sha256 data>`

## Task List
- [ ] ipfs-bot
- [ ] crawler (with webapi discovery support)
- [ ] monitor
- [ ] webapi (discovery only)
- [ ] setup nat streams (init container maybe?)
- [ ] exporter-vm
- [ ] exporter-pg
- [ ] grafana
- [ ] basic website
- [ ] automated reports
- [ ] automated reports on website