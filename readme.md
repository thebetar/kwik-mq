# Introduction

Kwik MQ is a basic messaging queue services written in GoLang

Kwik aims to be a simple messaging queue services completely written in Go, allowing a fast and flexible messaging queue which does not require a lot of system resources.

Kwik is the Dutch translation for the word Mercury, Mercury is the name of the Roman god of Messengers (https://en.wikipedia.org/wiki/Mercury_(mythology)).

# How to use

Kwik MQ can be used both using Docker and ran as a binary.

## Run using docker

To run the project using docker you can build the docker container yourself using

```
docker build . -t kwik-mw
```

or you can use docker compose which will automatically set it up to run on port 10526, this can be done by running

```
docker compose up
```

## Run using binary

You can run the `kwik-mq` binary to also run the project locally.
If you are not running Linux on x86 architecture you first need to run `make build` to recreate the binary for your own architecture.
