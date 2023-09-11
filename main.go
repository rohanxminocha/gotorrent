package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rohnxminocha/gotorrent/client"
)

func initializeLogger() *os.File {
	const logFileName string = "logs.txt"

	f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
	return f
}

func printHelp() {
	fmt.Println("exit: exits program")
	fmt.Println("print: prints all torrents")
	fmt.Println("add <file path>: adds torrent from <file path>")
	fmt.Println("remove <prefix>: removes first torrent with prefix <prefix>")
	fmt.Println("start <prefix>: starts downloading the torrent with prefix <prefix>")
	fmt.Println("stop <prefix>: stops downloading the torrent with prefix <prefix>")
}

func runClientCrudCommand(input []string, f func(string) error) {
	if len(input) != 2 {
		fmt.Println("not enough input arguments")
		return
	}

	err := f(input[1])
	if err != nil {
		fmt.Println(err)
	}
}

func runClient() {
	log.Println("Program bittorrent-client-go has started...")
	fmt.Println("Type 'help' for valid commands")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	torrentClient := client.New()
	for {
		scanner.Scan()
		input := strings.Fields(scanner.Text())
		if len(input) == 0 {
			continue
		}

		cmd := input[0]
		switch cmd {
		case "exit":
			os.Exit(0)
		case "help":
			printHelp()
		case "print":
			torrentClient.ShowTorrents()
		case "add":
			runClientCrudCommand(input, torrentClient.AddTorrent)
		case "remove":
			runClientCrudCommand(input, torrentClient.RemoveTorrent)
		case "start":
			runClientCrudCommand(input, torrentClient.StartTorrent)
		case "stop":
			runClientCrudCommand(input, torrentClient.StopTorrent)
		}

		fmt.Println()
	}
}

func main() {
	f := initializeLogger()
	defer f.Close()

	runClient()
}
