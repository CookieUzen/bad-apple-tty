# Bad-Apple-TTY

A program that renders a video in the terminal using gocv and ansi codes.

**Features:**
- true color support
- terminal size scaling
- frame rate/frame skipping (to keep up with a video)

## Modes

- tty: using black and white + 2 terminal blocks for 1 pixel
- tty_subsample: tty, but vertical pixels are subsampled into 1
- unicode: using black and white + half blocks for 2 pixels per character (this runs in tty)
- truecolor: unicode + true color 

## Install

This program relies on the [gocv library](https://gocv.io).
The issue is the gocv library installation process is a minor pain, so I would recommend running this binary in a docker container.

If by any chance you managed to install gocv, build the program with:

```bash
go build .
```

Else, run the docker container like so:

```bash
docker run -v $(pwd):$(pwd) -w $(pwd) -it ghcr.io/cookieuzen/bad-apple-tty /bad-apple-tty
```

Or put it in an alias:

```bash
alias bad-apple-tty="docker run -v $(pwd):$(pwd) -w $(pwd) -it ghcr.io/cookieuzen/bad-apple-tty /bad-apple-tty"
```

