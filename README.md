# ByProxy (byp)

Environment by proxy

## Installation

```
> dep ensure
> go install ./cmd/...
```

## Example config

```yaml
# ~/.byp/config.yaml

proxies:
- name: foobar

servers:
- name: example-server
  base-url: https://example.com

environments:
- name: demo
  mappings:
    foobar: example-server
```

Creates two proxies:

- `foobar.localhost` "foobar" proxy pointed to current environment
- `foobar-demo.localhsot` "foobar" proxy pointed to "demo" environment 

## Usage

Start server

```
> byp start
Starting ByProxy server
```

Server status

```
> byp status
Mappings:
foobar ->
Environments:
- demo
```

Change environment

```
> byp use demo
Mappings:
foobar -> demo
Environments:
- demo
```

Change environment for specific proxies

```
> byp use demo foobar
Mappings:
foobar -> demo
Environments:
- demo
```

Stop server

```
> byp stop
Stopping ByProxy server
```
