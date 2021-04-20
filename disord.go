package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
)

var client struct {
	s  *state.State
	ch discord.ChannelID
}

func main() {
	var err error
	client.s, err = state.New(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}
	err = client.s.Open()
	if err != nil {
		log.Fatalln(err)
	}
	sc := bufio.NewScanner(os.Stdin)
	input := make(chan string)
	events, _ := client.s.ChanFor(func(_ interface{}) bool { return true })
	go func() {
		for sc.Scan() {
			input <- sc.Text()
		}
	}()
	for {
		select {
		case line := <-input:
			handleInput(line)
		case ev := <-events:
			handleEvent(ev)
		}
	}
}

func handleInput(line string) {
	if line == ":q" || line == "\x05" {
		return
	}
	if strings.HasPrefix(line, ":") {
		splat := strings.Fields(line)
		cmd := line[1]
		switch cmd {
		case 'c':
			n, err := strconv.ParseInt(splat[1], 10, 64)
			if err != nil {
				log.Println(err)
				return
			}
			chid := discord.ChannelID(n)
			ch, err := client.s.Channel(chid)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println("join #" + ch.Name)
			client.ch = discord.ChannelID(n)
		}
		return
	}
	_, err := client.s.SendText(client.ch, line)
	if err != nil {
		log.Println(err)
	}
}

func handleEvent(ev interface{}) {
	switch ev := ev.(type) {
	case *gateway.MessageCreateEvent:
		if ev.ChannelID != client.ch {
			return
		}
		printMsg(m)
	}
}

func printMsg(m discord.Message) {
	lines := strings.Split(m.Content, "\n")
	for _, line := range lines {
		fmt.Printf("%s → %s\n", m.Author.Username, line)
	}
}
