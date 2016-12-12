// Copyright 2013, 2014, 2016 Tom Most
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
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/go-log/log"
	irc "github.com/fluffle/goirc/client"
	"github.com/stvp/go-toml-config"
	"io"
	"math/rand"
	"net"
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

// ircConfig creates an IRC configuration based on global configuration flags.
func ircConfig() (*irc.Config, error) {
	var host string
	var err error
	host, _, err = net.SplitHostPort(*ircServer)
	if err != nil {
		return nil, err
	}
	cfg := irc.NewConfig(*ircNick, *ircNick, *ircName)
	cfg.Server = *ircServer
	if *ircSSL {
		cfg.SSL = true
		cfg.SSLConfig = &tls.Config{ServerName: host}
	}
	return cfg, nil
}

// yearEndGreeting returns an IRC-formatted string which wishes root-vegetable
// related good fortune.
func yearEndGreeting() string {
	if rand.Float64() < 0.8 {
		// "carrot" is orange (7).
		return "And a \x037carrot\x03 New Year!"
	} else {
		// "turnip" is purple (6).
		return "And a \x036turnip\x03 New Year!"
	}
}

func main() {
	configFile := flag.String("config", "config.toml", "Config file")
	flag.Parse()

	config.Parse(*configFile)
	if err := seed(); err != nil {
		log.Fatalf("Unable to seed the PRNG: %s", err)
	}

	carrotFacts, err := loadFacts(*carrotDB)
	if err != nil {
		log.Fatalf("Unable to read facts from '%s': %s", *carrotDB, err)
	}
	if carrotFacts == nil || len(carrotFacts) < 1 {
		log.Fatalf("No facts available")
	}
	log.Printf("Loaded %d carrot facts", len(carrotFacts))

	turnipFacts, err := loadFacts(*turnipDB)
	if err != nil {
		log.Fatalf("Unable to read facts from '%s': %s", *turnipDB, err)
	}
	if turnipFacts == nil || len(turnipFacts) < 1 {
		log.Fatalf("No facts available")
	}
	log.Printf("Loaded %d turnip facts", len(turnipFacts))

	// Index into turnipFacts, the next fact to output.
	turnipIndex := 0

	cfg, err := ircConfig()
	if err != nil {
		log.Fatalf("Failed to configure IRC: %s", err)
	}

	ic := irc.Client(cfg)

	ic.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Joining %s", *ircChannel)
			conn.Join(*ircChannel)
		})

	ic.HandleFunc("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
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
		} else if strings.Contains(line.Args[1], "Merry") && strings.Contains(line.Args[1], "Christmas") {
			conn.Privmsg(channel, yearEndGreeting())
		}
	})

	quit := make(chan bool)
	ic.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	log.Printf("Connecting to %s", ic.Config().Server)
	if err := ic.Connect(); err != nil {
		log.Fatalf("Unable to connect: %s", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

main:
	for {
		select {
		case <-interrupt:
			log.Printf("Interrupted: shutting down")
			ic.Quit("Carrot be with you!")
		case <-quit:
			log.Printf("Disconnected")
			break main
		}
	}
}
