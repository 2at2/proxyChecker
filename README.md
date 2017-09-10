# Proxy checker

Checks proxy list and builds anonymous alive list.

## Release
Run command:
```
make deps
make release
```

## Target
Listens port and responds information about request.

Params:
* --port - 80 by default

Command:
```
release/target --port 8080
```

## Checker

Params:
* --target - Target host. For example: 127.0.0.1:8080
* --file - File of proxies
* --result - Result file

Example file of proxies:
```
192.168.0.1:9091
192.168.0.2:80
```

Command:
```
release/checker --target 127.0.0.1:8080 --file pathToFile --result pathToFile
```