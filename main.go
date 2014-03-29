// Copyright 2013, 2014 Tom Most
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"github.com/stvp/go-toml-config"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
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

// loadFacts obtains facts from the named database file in JSON or TXT format.
//
// The JSON format is a list of objects with a "text" key, containing the fact
// text, and an "id" key, an integer identifier.  The file name must end with
// ".json".
//
// The TXT format is one fact per line.  Line numbers are used as identifiers.
// The file name must end with ".txt".
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

// seed feeds the Go PRNG a cryptographically random number so we don't always
// choose facts in the same order.
func seed() (err error) {
	var seed int64
	err = binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	if err == nil {
		rand.Seed(seed)
	}
	return
}

// choose picks a random fact from the collection.
func choose(facts []Fact) (fact *Fact) {
	if facts == nil {
		return nil
	}
	i := rand.Intn(len(facts))
	return &facts[i]
}

func main() {
	configFile := flag.String("config", "config.toml", "Config file")
	flag.Parse()

	config.Parse(*configFile)
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
			if turnipIndex == len(turnipFacts) {
				turnipIndex = 0
			}
			conn.Privmsg(channel, turnip.Text)
		} else if strings.HasPrefix(line.Args[1], ".carro") {
			count := strings.Count(line.Args[1], "o")
			if count > 61 {
				count = 61 // max which fit
			}
			msg := strings.Repeat("CARROT ", count)
			conn.Privmsg(channel, msg[0:len(msg)-1])
		}
	})

	quit := make(chan bool)
	ic.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	if err := ic.Connect(*ircServer); err != nil {
		log.Fatalf("Unable to connect: %s\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

main:
	for {
		select {
		case <-interrupt:
			ic.Quit("Carrot be with you!")
		case <-quit:
			log.Printf("Disconnected")
			break main
		}
	}
}
