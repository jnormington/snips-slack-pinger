# Snips slack pinger

As the name implies it utilises Snips for the voice/intent processing and listens on a specific intent queue for messages for the slack intent. When a match is found it will attempt to lookup the user/channel input and post a message. It is mainly used for pinging users who haven't turned up to standup and ping the relevant channel.

It uses the dynamic entity injection so that you don't need to put personal names on the snips slot in the console as its performs this action every 7 hours to ensure any new entries exist for the next standup.

Unfortunately this won't work in the Snip ecosystem of linking actions :( but it is very easy to run and deploy to your raspberry pi

## Download prebuilt binary

[Darwin x64](https://github.com/jnormington/snips-slack-pinger/releases/download/v0.0.1/ssp-0.0.1-darwinx64)

[Linux arm](https://github.com/jnormington/snips-slack-pinger/releases/download/v0.0.1/ssp-0.0.1-linux-arm)

[Linux x86](https://github.com/jnormington/snips-slack-pinger/releases/download/v0.0.1/ssp-0.0.1-linux386)

[Linux x64](https://github.com/jnormington/snips-slack-pinger/releases/download/v0.0.1/ssp-0.0.1-linux64)


## Generate example config

```sh
./ssp-* -generate-config > config.json
```

Update the config options relevant to you. Then you are ready to run the program

## Run

To run it in dry-run mode - this will NOT message anybody in slack, and will just output in the log and prefix the message with [DRYRUN] so you know who it would have messaged and the ID for that user.

```sh
./ssp-* -config config.json -dry-run
```

To run for real just remove the `-dry-run` switch from the command

# Build from source

It is expected that you have installed golang on the relevant device 
(or have go installed on a machine that you can make a build for the target device)

- Clone this repository
- make build

Make build defaults to linux arm build which is for raspberry and other linux arm devices), and assumes that you are not building it on the target device (ie raspberry pi)

If you want to compile for other devices check out the [go build syslist](https://github.com/golang/go/blob/master/src/go/build/syslist.go)

```
   GOOS=linux GOARCH=amd64 make build
```

If you are building from source on the target device you can just use `go build`
