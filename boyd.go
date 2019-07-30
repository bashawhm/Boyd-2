package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	irc "github.com/thoj/go-ircevent"
)

var quoteList []string
var randGen *rand.Rand

func filter(array []string, f func(string) bool) []string {
	filteredArray := make([]string, 0)
	for _, str := range array {
		if f(str) {
			filteredArray = append(filteredArray, str)
		}
	}
	return filteredArray
}

func getQuote() string {
	if len(quoteList) == 0 {
		return "No quotes found..."
	}
	return quoteList[randGen.Int()%len(quoteList)]
}

func getSearchQuote(search string) string {
	if len(quoteList) == 0 {
		return "No quotes found..."
	}

	filteredQuotes := filter(quoteList, func(str string) bool {
		return strings.Contains(str, search)
	})

	if len(filteredQuotes) == 0 {
		return "No quotes found with that search query..."
	}

	return filteredQuotes[randGen.Int()%len(filteredQuotes)]
}

func loadQuotes(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	fin := bufio.NewScanner(bufio.NewReader(file))
	fin.Split(bufio.ScanLines)
	for fin.Scan() {
		quoteList = append(quoteList, fin.Text())
	}
	file.Close()
}

func writeAllQuotes(fout *bufio.Writer) {
	for i := 0; i < len(quoteList); i++ {
		fout.WriteString(quoteList[i] + "\n")
	}
	fout.Flush()
}

func writeQuote(fout *bufio.Writer, quote string) {
	fout.WriteString(quote + "\n")
	fout.Flush()
}

func stripPrefix(prefix, data string) string {
	var res string
	for i := 0; i < len(data); i++ {
		if strings.HasPrefix(data[:i], prefix) {
			res = data[i:]
			break
		}
	}
	return res
}

func main() {
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	roomNames := []string{"#testit"}
	botName := "boyd_bot"
	serverNamePort := "irc.freenode.net:6667"

	loadQuotes("./quotes.txt")

	file, err := os.Create("./quotes.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fout := bufio.NewWriter(file) //I'm the big dumb so until I figure out a better way, let's just live with this
	writeAllQuotes(fout)

	conn := irc.IRC(botName, botName)
	err = conn.Connect(serverNamePort)
	if err != nil {
		fmt.Println("failed to connect")
		return
	}

	conn.AddCallback("001", func(e *irc.Event) {
		for i := 0; i < len(roomNames); i++ {
			conn.Join(roomNames[i])
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		msg := e.Message()
		if strings.HasPrefix(msg, "!quoteadd ") {
			res := stripPrefix("!quoteadd ", msg)
			fmt.Println("Adding quote: " + res)
			writeQuote(fout, res)
			quoteList = append(quoteList, res)
			conn.Privmsg(e.Arguments[0], "Added!")
		} else if strings.HasPrefix(msg, "!quote") {
			res := stripPrefix("!quote", msg)
			if len(res) != 0 {
				searchMsg := stripPrefix(" ", res)
				ret := getSearchQuote(searchMsg)
				conn.Privmsg(e.Arguments[0], ret)
				return
			}
			conn.Privmsg(e.Arguments[0], getQuote())
		} else if strings.Contains(msg, botName) {
			conn.Privmsg(e.Arguments[0], (buildsentence(5, 5)))
		}
	})

	conn.Loop()
}
