## Socket Activation for Caddy

Useful in deployments where sockets are passed by systemd.

### Usage

#### Caddy Configuration

```Caddyfile
localhost:443 {
    bind socket-activation/https
    respond "ok"
}
```

Using socket activation requires every site block to have explicit **bind**, it can be used in a snippet to avoid repetition.

### Usage

#### Commandline

```bash
systemd-socket-activate -l 443 --fdname=https systemd-socket-activate --datagram -l 443 --fdname=https:https caddy run
```
