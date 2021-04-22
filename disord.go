package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"golang.org/x/term"
)

var client struct {
	s  *state.State
	ch discord.ChannelID
}

var out io.Writer
var oldState *term.State

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
	input := make(chan string)
	errs := make(chan error)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalln(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
		c := struct {
			io.Reader
			io.Writer
		}{os.Stdin, os.Stdout}
		term := term.NewTerminal(c, "→ ")
		out = term
		log.SetOutput(term)
		go terminalInput(term, input, errs)
	} else {
		out = os.Stdout
		go stdinInput(input, errs)
	}
	events, _ := client.s.ChanFor(func(_ interface{}) bool { return true })
	for {
		select {
		case line := <-input:
			handleInput(line)
		case ev := <-events:
			handleEvent(ev)
		case err = <-errs:
			log.Fatalln(err)
		}
	}
}

func stdinInput(input chan string, errs chan error) {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		input <- sc.Text()
	}
	errs <- sc.Err()
}

func terminalInput(term *term.Terminal, input chan string, errs chan error) {
	for {
		line, err := term.ReadLine()
		if err != nil {
			errs <- err
			break
		}
		input <- line
	}
}

func handleInput(line string) {
	if strings.HasPrefix(line, ":") {
		splat := strings.Fields(line)
		cmd := splat[0][1:]
		handleCommand(cmd, splat[1:])
		return
	}
	_, err := client.s.SendText(client.ch, line)
	if err != nil {
		log.Println(err)
	}
}

func handleCommand(cmd string, args []string) {
	switch cmd {
	case "c":
		n, err := strconv.ParseInt(args[0], 10, 64)
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
		fmt.Fprintln(out, "join #"+ch.Name)
		client.ch = discord.ChannelID(n)
	case "m":
		n, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			log.Println(err)
			return
		}
		msgs, err := client.s.Client.Messages(client.ch, uint(n))
		if err != nil {
			log.Println(err)
			return
		}
		for i := len(msgs) - 1; i >= 0; i-- {
			printMsg(msgs[i])
		}
	}
	return
}

func handleEvent(ev interface{}) {
	switch ev := ev.(type) {
	case *gateway.MessageCreateEvent:
		if ev.ChannelID != client.ch {
			return
		}
		printMsg(ev.Message)
	}
}

func printMsg(m discord.Message) {
	lines := strings.Split(m.Content, "\n")
	for _, line := range lines {
		fmt.Fprintf(out, "%s → %s\n", m.Author.Username, line)
	}
}
