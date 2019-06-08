package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	termbox "github.com/nsf/termbox-go"
)

var piZero = `
######### Raspbery Pi Model Zero GPIO simulator ###########

 +| +| -|14|15|18| -|23|24| -|25|08|07|01| -|12| -|16|20|21
         ## ## ##    ## ##    ## ## ## ##    ##    ## ## ##
	 ## ## ##    ## ## ##    ## ## ##    ## ## ## ## ## ##
 +|02|03|04| -|17|27|22| +|10|09|11| -|00|05|06|13|19|16| -
`

var pin14 = []int{2, 9}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	lines := strings.Split(piZero, "\n")

	for row, l := range lines {
		for col, r := range l {
			termbox.SetCell(col, row, r, termbox.ColorWhite, termbox.ColorBlack)
		}
	}

	termbox.Flush()
	time.Sleep(2 * time.Second)

	termbox.SetCell(pin14[1], pin14[0], '#', termbox.ColorRed, termbox.ColorBlack)
	termbox.SetCell(pin14[1]+1, pin14[0], '#', termbox.ColorRed, termbox.ColorBlack)
	termbox.Flush()

	c := make(chan os.Signal, 1)
	signal.Notify(c)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal", s)
}
