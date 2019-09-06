package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"regexp"

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

var lastSearches map[string][]string
var searchOrder []string
const MAX_SEARCHES int = 5

func getSearchQuote(search string) string {
	if(lastSearches == nil) {
		lastSearches = make(map[string][]string)
	}

	sl, ok := lastSearches[search]
	if ok && len(sl) > 0 {
		fmt.Printf("Search found in lS, %d remaining\n", len(sl))
		ret := sl[0]
		if len(sl) == 1 {
			fmt.Println("lS emptied of search")
			delete(lastSearches, search)
			searchOrder = filter(searchOrder, func(s string) bool {
				return s != search
			})
		} else {
			lastSearches[search] = sl[1:]
			fmt.Printf("...now %d\n", len(sl) - 1)
		}
		return ret
	}

	if len(quoteList) == 0 {
		return "No quotes found..."
	}

	re, err := regexp.Compile(search)
	if err != nil {
		return "Error compiling pattern: " + err.Error()
	}

	filteredQuotes := filter(quoteList, func(str string) bool {
		return re.FindStringSubmatch(str) != nil
	})

	fmt.Printf("Fresh search made %d matches\n", len(filteredQuotes))

	if len(filteredQuotes) == 0 {
		return "No quotes found with that query..."
	}

	shuffled := make([]string, len(filteredQuotes))
	for idx, perm := range randGen.Perm(len(filteredQuotes)) {
		shuffled[perm] = filteredQuotes[idx]
	}

	ret := shuffled[0]
	shuffled = shuffled[1:]
	if len(shuffled) > 0 {
		fmt.Printf("Reshuffling %d matches into lS\n", len(shuffled))
		lastSearches[search] = shuffled
		searchOrder = append(searchOrder, search)
		if len(searchOrder) > MAX_SEARCHES {
			wm := len(searchOrder) - MAX_SEARCHES
			for _, v := range searchOrder[:wm] {
				delete(lastSearches, v)
			}
			searchOrder = searchOrder[wm:]
		}
		fmt.Printf("Current sO is %#v\n", searchOrder)
	}

	return ret
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
	roomNames := []string{"#test3b19763a92c"}
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
		fmt.Println("(joining)")
		for i := 0; i < len(roomNames); i++ {
			conn.Join(roomNames[i])
		}
	})

	conn.AddCallback("NOTICE", func(e *irc.Event) {
		fmt.Printf("%+v\n", e)
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		msg := e.Message()
		target := e.Arguments[0]
		if target[0] != '#' {
			target = e.Nick  // direct message
		}
		fmt.Printf("%v from %v to %v (target %v) === %v\n", e.Arguments, e.Nick, e.Arguments[0], target, msg)
		if strings.HasPrefix(msg, "!quoteadd ") {
			res := stripPrefix("!quoteadd ", msg)
			fmt.Println("Adding quote: " + res)
			writeQuote(fout, res)
			quoteList = append(quoteList, res)
			conn.Privmsg(target, "Added!")
		} else if strings.HasPrefix(msg, "!quote") {
			res := stripPrefix("!quote", msg)
			searchMsg := stripPrefix(" ", res)
			ret := getSearchQuote(searchMsg)
			conn.Privmsg(target, ret)
		} else if strings.Contains(msg, botName) {
			conn.Privmsg(target, (buildsentence(5, 5)))
		}
	})

	conn.Loop()
}
