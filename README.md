## Socket Activation for Caddy (WIP)

Useful in deployments where sockets are passed by systemd.

### Usage

#### Caddy Configuration

```Caddyfile
localhost:443 {
    bind socket-activation/localhost
    respond "ok"
}
```

Using socket activation requires every site block to have explicit **bind**, it can be used in a snippet to avoid repetition.

### Usage

#### Commandline

```bash
systemd-socket-activate -l 443 --fdname=localhost systemd-socket-activate --datagram -l 443 --fdname=localhost:localhost_UDP caddy run
```

### Credits

FD to sockets mapping comes from [go-systemd](https://github.com/coreos/go-systemd/).
