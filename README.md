# [kudoer.com](https://kudoer.com)

It's a site that allows you to review, or "give kudos" to anything you please.
Books, movies, places, mathematical concepts, whatever. You can follow your
friends accounts and see their reviews.

## compile

A `go` compiler is required to compile this application. Check `go.mod` for the
oldest supported version of [go](https://go.dev/). Then run `make` to compile
the project.

## usage

Run with the `-help` flag for current options. Email settings are for the
"forgot password" feature. If you're actually trying to host this make sure you
put it behind a proxy (caddy, nginx, openbsd httpd, apache, etc) as the
application does not serve https by itself.

## license

GNU AGPL version 3 or later, see LICENSE.

Emoji are provided by [Noto Color Emoji](https://github.com/googlefonts/noto-emoji) and used via the [SIL Open Font License, version 1.1](https://github.com/googlefonts/noto-emoji/blob/main/fonts/LICENSE).
