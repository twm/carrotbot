=========
carrotbot
=========

An IRC bot which responds to the ".carrot" command with a random carrot fact,
courtesy of `@RealCarrotFacts <https://twitter.com/RealCarrotFacts>`_.

Quotes are loaded from a JSON file as output by `twistorpy
<https://github.com/fisadev/twistorpy>`_, by default in ``facts.json``.  To
generate it::

    $ python2.7 twistorpy.py RealCarrotFacts facts.json

Updates can be done thereafter with the same command.

Configuration is done via ``config.toml``, a file in the TOML (INI-like)
format.  An example::

    [irc]
    server = irc.freenode.net:7000
    ssl = true
    user = fontofwisdom
    password = supersekrit
    nick = fontofwisdom
    name = Font of Wisdom
    channel = #your-channel

    [facts]
    db = facts.json

carrotbot has only been tested against Freenode's servers.
