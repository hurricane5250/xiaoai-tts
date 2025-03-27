package main

import (
	"bufio"
	"fmt"
	"os"

	xiaoaitts "github.com/hurricane5250/xiaoai-tts"
)

func main() {
	x, err := xiaoaitts.New("xxxx", "xxxx")
	if err != nil {
		return
	}

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Please Input Text")
		text, _ := input.ReadString('\n')
		x.Say(text)
	}
}
