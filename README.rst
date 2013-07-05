=========
carrotbot
=========

An IRC bot which espouses root-vegetable related "facts".  This version
supports the ".carrot" and ".turnip" commands.

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

carrotbot has only been tested against Freenode's servers.
