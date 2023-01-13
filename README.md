# Coinche

## Description

Simple Coinche game to play with friends remotely in a browser

## Initialize go modules files

```
go mod init github.com/desbouis/coinche
go mod tidy
```

## Build container image with podman

```
podman build -t coinche .
```

## Run app with systemd+podman

```
systemctl --user start podman-kube@$(systemd-escape </path/to>/coinche/service.yml).service
```

And then got to http://localhost:8080/coinche/
