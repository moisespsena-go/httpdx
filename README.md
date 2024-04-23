# httpdx
Provides HTTP reverse proxy and TCP proxies by websocket.

Run `httpdx -h` for usage detail.


## Create Config

Create config file from template.

Runs `httpdx create-config -h` to usage.

```
Usage:
httpdx create-config [OPTIONS] ARG...

Options:
  -out string
        The output file (default "-")
  -server-addr string
        The httpdx server addr (default ":80")
  -server-url string
        The httpdx server url (default "http://127.0.0.1:80")
```

### Config File Example

See to `httpdx create-config` to expended doc.

```yml
client:
  server_url: "{{.ServerUrl}}"

  # default authentication configuration
  auth:
    user: my-user
    password: 123
    disabled: false

  routes:
    - name: ssh
      local_addr: :25000

      # authentication configuration for this route.
      # If not set, uses default auth configuration
      auth:
        user: my-user
        password: 123
        disabled: false

server:
  addr: "{{.ServerAddr}}"

  # not found HTML file to handles not found error.
  # If not set, uses default not found handler message.
  not_found: "my_not_found.html"

  # if is true, disables not found handles
  not_found_disabled: false

  tcp_sockets:
    # timeouts is in seconds (default is 5s).

    handshake_timeout: 5
    dial_timeout: 5
    write_timeout: 5

    compression_enabled: false

    # default authentication configuration
    auth:
      user: my-user
      password: 123
      disabled: false

    routes:
      ssh:
        addr: localhost:22
        disabled: false

        # authentication configuration for this route.
        # If not set, uses default auth configuration
        auth:
          user: my-user
          password: 123
          disabled: false

  http:
    routes:
      /:
        addr: 127.0.0.1:80
        disabled: false

      /admin/:
        addr: 127.0.0.1:81

      /pth:
        addr: 127.0.0.1:81

      # static files serving
      /static/:
        # if requests /static/f.txt, responds content from ./static_files/f.txt
        dir: ./static_files
        # if set, overrides real file path. Default value is "%[1]s"
        # This formatter haves 3 values:
        # 1. The requested path
        # 2. The DIR
        # 3. The route PATH
        # Examples with request path is /static/f.txt:
        # - a/b/%[1]s: responds content from ./static_files/a/b/f.txt
        # - %[2]s/%[1]s: responds content from ./static_files/static_files/a/b/f.txt
        # - %[3]s/%[1]s: responds content from ./static_files/static/a/b/f.txt
        path_override: "%[3]s/%[1]s"

      # when path_strip is true, proxifies /my-dir as / to destination and pass '/my-dir'
      # into request header 'path_header' (default is 'X-Forwarded-Prefix')
      /my-dir:
        addr: 127.0.0.1:80
        path_strip: true
        path_header: X-Forwarded-Prefix
        disabled: false
```

## Server

Starts the server.


Runs `httpdx server -h` or `httpdx server [OPTIONS ARG];;;`.

Runs `httpdx server -h` to usage:

```
Usage:
httpdx server [OPTIONS]

Options:
  -config string
        YAML configuration file (default "___8httpdx.yml")
  -addr string
        The server Address
```


`httpdx` or `httpdx -config ./httpdx.yml`

if requests contains header `X-Httpdx-Handle-Fallback: false`, disables Not Found handlers.

## Client

Starts the cliente to dispose remote tcp_sockets into local addr.

Runs `httpdx client -h` to usage.

```
Usage:
httpdx_client [OPTIONS] ARG...

Options:
  -config string
        YAML configuration file (default "___httpdx_client.yml")
  -server-url string
        The httpdx server url
```

Pass services as args `NAME:LOCAL_ADDR`.

- `httpdx client`, or 
- `httpdx -config ./httpdx.yml client`, or
- `httpdx client ssh:localhost:26000 other:localhost:26001`, or
- `httpdx -config ./httpdx.yml client ssh:localhost:26000 other:localhost:26001` 
