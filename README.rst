=========
carrotbot
=========

An IRC bot which espouses root-vegetable related "facts".  This version
supports the ".carrot" and ".turnip" commands.

Building
--------

Carrotbot is written in Go using the `GoIRC library <https://github.com/fluffle/goirc>`.
Dependencies are managed with `Glide <https://glide.sh>` and the ``vendor`` directory is checked in.
To build you will need a functioning Go install (I use 1.7, but older versions will likely work) and ``GOPATH`` set.
Then::

    go build .

Quote Databases
---------------

Two formats are supported for quote databases: a newline-delimited UTF-8 text
file, or JSON.  The filetype of the database is indicated by the filename
extension, either ".txt" or ".json".

The JSON format is compatible with the output of `twistorpy
<https://github.com/fisadev/twistorpy>`_.  It is an array of object literals
with at least "id" (integer) and "text" (string) properties.  For example, to
generate one for `@RealCarrotFacts <https://twitter.com/RealCarrotFacts>`::

    $ python2.7 twistorpy.py RealCarrotFacts carrots.json

Updates can be done thereafter with the same command.

Configuration
-------------

Configuration is done via a file in the TOML format, by default
``config.toml``, but configurable via the ``-config`` flag.  An example::

    [irc]
    server = "irc.freenode.net:7000"
    ssl = true
    user = "fontofwisdom"
    password = "supersekrit"
    nick = "fontofwisdom"
    name = "Font of Wisdom"
    channel = "#your-channel"

    [facts]
    carrots = "carrots.json"
    turnips = "turnips.txt"

    [logging]
    # log to stderr rather than the systemd journal
    journal = false
    # debug logs include everything said in the channel
    debug = true

carrotbot has only been tested against Freenode's servers.
