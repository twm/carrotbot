package main

import (
	irc "github.com/fluffle/goirc/client"
	"github.com/stvp/go-toml-config"
	"log"
)

var factFile = config.String("factfile", "facts.json")

var (
	ircServer  = config.String("irc.server", "irc.freenode.net:7000")
	ircSSL     = config.Bool("irc.ssl", true)
	ircNick    = config.String("irc.nick", "carrotfacts")
	ircName    = config.String("irc.name", "")
	ircUser    = config.String("irc.user", "carrotfacts")
	ircPass    = config.String("irc.password", "")
	ircChannel = config.String("irc.channel", "#carrotfacts-test")
)

func main() {
	config.Parse("config.toml")

	ic := irc.SimpleClient(*ircNick, *ircNick, *ircName)
	ic.SSL = *ircSSL

	ic.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Connected\n")
			conn.Join(*ircChannel)
		})

	ic.AddHandler("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
		log.Printf("Got message: %s\n", line)
		if line.Args[1] == ".carrot" {
			// TODO: Pick a random message
			conn.Notice(line.Args[0], "I can has carrot?")
		}
	})

	quit := make(chan bool)
	ic.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	if err := ic.Connect(*ircServer); err != nil {
		log.Fatalf("Unable to connect: %s\n", err)
		return
	}

	<-quit
}
