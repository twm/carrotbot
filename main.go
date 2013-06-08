package main

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	irc "github.com/fluffle/goirc/client"
	"github.com/stvp/go-toml-config"
	"io"
	"log"
	"math/rand"
	"os"
)

var factFile = config.String("factfile", "facts.json")

// IRC configuration parameters
var (
	ircServer  = config.String("irc.server", "irc.freenode.net:7000")
	ircSSL     = config.Bool("irc.ssl", true)
	ircNick    = config.String("irc.nick", "carrotfacts")
	ircName    = config.String("irc.name", "")
	ircUser    = config.String("irc.user", "carrotfacts")
	ircPass    = config.String("irc.password", "")
	ircChannel = config.String("irc.channel", "#carrotfacts-test")
)

// Fact database
var (
	factDB = config.String("facts.db", "facts.json")
)

type CarrotFact struct {
	Text string `json:"text"`
	Id   int64  `json:"id"`
}

func loadFacts(fn string) (facts []CarrotFact, err error) {
	var f io.Reader
	f, err = os.Open(fn)
	if err != nil {
		return
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(&facts)
	return
}

// The Go PNRG must be seeded or we'll always give facts in the same order.
func seed() (err error) {
	var seed int64
	err = binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	if err == nil {
		rand.Seed(seed)
	}
	return
}

// Pick a random fact from the collection.
func choose(facts []CarrotFact) (fact *CarrotFact) {
	if facts == nil {
		return nil
	}
	i := rand.Intn(len(facts))
	return &facts[i]
}

func main() {
	config.Parse("config.toml")
	if err := seed(); err != nil {
		log.Fatalf("Unable to seed the PRNG: %s\n", err)
	}

	facts, err := loadFacts(*factDB)
	if err != nil {
		log.Fatalf("Unable to read facts from '%s': %s\n", *factDB, err)
	}
	if facts == nil || len(facts) < 1 {
		log.Fatalf("No facts available\n")
	}
	log.Printf("Loaded %d facts\n", len(facts))

	ic := irc.SimpleClient(*ircNick, *ircNick, *ircName)
	ic.SSL = *ircSSL

	ic.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			conn.Join(*ircChannel)
		})

	ic.AddHandler("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
		if line.Args[1] == ".carrot" {
			fact := choose(facts)
			conn.Notice(line.Args[0], (*fact).Text)
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
