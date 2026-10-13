# Local DNS for Shizzi hotspots

The Local DNS section in Settings can assign arbitrary IPv4 addresses to local
hostnames, without changing DNS settings on each connected hotspot client.

## Example

Connect a Windows Docker Desktop laptop to the Samsung hotspot. If its hotspot
address is `10.201.244.112`:

1. Add `photocraft` -> `10.201.244.112`.
2. Add `pdf` -> `10.201.244.112`.
3. Run a reverse proxy such as Caddy on that laptop's port 80:

```caddyfile
http://photocraft {
    reverse_proxy host.docker.internal:8090
}
http://pdf {
    reverse_proxy host.docker.internal:8080
}
```

Caddy must run in Docker Desktop for `host.docker.internal` to resolve.
Allow inbound TCP port 80 on Windows for the hotspot network.

From a second hotspot client, try `nslookup photocraft`, and then
`http://photocraft` in the browser. A bare name might be treated as a search,
so include `http://`.

## Limitations

- Local DNS maps hostnames to addresses only; a reverse proxy must select
  which container/port handles the HTTP request.
- IPv4 A records only. AAAA queries for local names yield NOERROR/NODATA.
- Hotspot clients with custom encrypted DNS may bypass this resolver.
- Android's tethering DNS forwarder must use the active Shizzi TUN network;
  behavior needs verification on each OEM/Android combination.
- A laptop DHCP address may change when it reconnects. Update the DNS entry.
- Local DNS editing uses DataStore, and active sessions receive new snapshots.
  Invalid entries are rejected before replacing the current snapshot.
- Rebuilding from source requires Go and Android toolchains as in the Build
  workflow. Do not treat this as verified until the Go tests and Android CI pass.
