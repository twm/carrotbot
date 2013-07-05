package main

import (
	"bufio"
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/stvp/go-toml-config"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
)

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

// Fact databases
var (
	carrotDB = config.String("facts.carrots", "carrots.json")
	turnipDB = config.String("facts.turnips", "turnips.txt")
)

type Fact struct {
	Text string `json:"text"`
	Id   int64  `json:"id"`
}

// Load facts from a database in JSON or TXT format.
//
// The JSON format is a list of objects with a "text" key, containing the fact
// text, and an "id" key, an integer identifier.
//
// The TXT format is one fact per line.  Line numbers are used as identifiers.
func loadFacts(fn string) (facts []Fact, err error) {
	var f io.Reader
	f, err = os.Open(fn)
	if err != nil {
		return
	}
	if strings.HasSuffix(fn, ".json") {
		dec := json.NewDecoder(f)
		err = dec.Decode(&facts)
	} else if strings.HasSuffix(fn, ".txt") {
		facts, err = loadLines(f)
	} else {
		err = fmt.Errorf("filename %q has unknown extension (not .json or .txt)", fn)
	}
	return
}

// Read in a newline-delimited fact file.
func loadLines(f io.Reader) ([]Fact, error) {
	scanner := bufio.NewScanner(f)
	facts := make([]Fact, 0)
	var line int64 = 0
	for scanner.Scan() {
		facts = append(facts, Fact{
			Id:   line,
			Text: scanner.Text(),
		})
		line++
	}
	return facts, scanner.Err()
}

// The Go PRNG must be seeded or we'll always give facts in the same order.
func seed() (err error) {
	var seed int64
	err = binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	if err == nil {
		rand.Seed(seed)
	}
	return
}

// Pick a random fact from the collection.
func choose(facts []Fact) (fact *Fact) {
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

	carrotFacts, err := loadFacts(*carrotDB)
	if err != nil {
		log.Fatalf("Unable to read facts from '%s': %s\n", *carrotDB, err)
	}
	if carrotFacts == nil || len(carrotFacts) < 1 {
		log.Fatalf("No facts available\n")
	}
	log.Printf("Loaded %d carrot facts\n", len(carrotFacts))

	turnipFacts, err := loadFacts(*turnipDB)
	if err != nil {
		log.Fatalf("Unable to read facts from '%s': %s\n", *turnipDB, err)
	}
	if turnipFacts == nil || len(turnipFacts) < 1 {
		log.Fatalf("No facts available\n")
	}
	log.Printf("Loaded %d turnip facts\n", len(turnipFacts))

	// Index into turnipFacts, the next fact to output.
	turnipIndex := 0

	ic := irc.SimpleClient(*ircNick, *ircNick, *ircName)
	ic.SSL = *ircSSL

	ic.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			conn.Join(*ircChannel)
		})

	ic.AddHandler("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
		channel := line.Args[0]
		if line.Args[1] == ".carrot" {
			carrot := choose(carrotFacts)
			conn.Privmsg(channel, (*carrot).Text)
		} else if line.Args[1] == ".turnip" {
			turnip := turnipFacts[turnipIndex]
			turnipIndex++
			if turnipIndex == len(turnipFacts) { turnipIndex = 0 }
			conn.Privmsg(channel, turnip.Text)
		} else if strings.HasPrefix(line.Args[1], ".carroop") {
			conn.Privmsg(channel, "CARROT CARROT CARROT CARROT")
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
