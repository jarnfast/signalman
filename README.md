# signalman

Small application that exposes a HTTP endpoint which can be used to send (UNIX) signals to a child process.

Wrapping the main container's application with signalman allows sidecar containers to control behavior using HTTP requests.

E.g. very usefull when doing lift-and-shift of legacy applications into Kubernetes and the application supports reloading configurations when receiving a SIGHUP signal.

## Security disclaimer

The exposed HTTP API is not served with TLS nor does it require any form of authentication.

## Getting signalman

Download the latest binary from the release page or build it yourself (see [Building signalman](#building-signalman))

## Usage

Simply start signalman with the command (and args) that should be wrapped:

```sh
$ signalman <command> [<args> ...]
```

Examples:

```sh
$ signalman /some/app arg1 arg2

$ signalman sleep 60
```

### HTTP API

| Method | Path | Description |
| :----- | :--- | :---------- |
| `GET` | `/status` | Return status with version info |
| `POST` | `/term` | Sends signal `TERM` to wrapped process. If the process doesn't terminate within `SIGNALMAN_TERM_TIMEOUT` seconds a `KILL` is sent |
| `POST` | `/kill` | Sends signal `KILL` to the wrapped process |
| `POST` | `/signal/<num>` | Sends signal `<num>` to the wrapped process |

Example:

```sh
$ curl -X POST http://localhost:30000/term
```

### Configuration

signalman supports a few configuration options using environment variables:

| Name | Description | Default (if not set) |
| :--- | :---------- | :------ |
| `SIGNALMAN_LISTEN_ADDRESS` | Address and port signalman will listen for HTTP requests on| `localhost:30000` |
| `SIGNALMAN_TERM_TIMEOUT` |  Duration in seconds to wait for process to response to TERM signal | `10` |
| `SIGNALMAN_LOG_LEVEL` | Log level (debug/info/warn/error/dpanic/panic/fatal) | `info`
| `SIGNALMAN_LOG_CONFIGFILE` | Path to file containing [:zap: uber-go/zap](https://github.com/uber-go/zap) logger config. See `testdata/example-log-config.json` for example | |

## Building signalman

signalman can be built using `make`

```sh
$ make build
```

.. or cross-built:

```sh
$ make PLATFORM=linux ARCH=arm GOARM=6 xbuild
```

