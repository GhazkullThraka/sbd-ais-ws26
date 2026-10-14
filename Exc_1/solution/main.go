package main

import "fmt"

type Message struct {
	say1 string
	say2 string
}

func (m Message) Print() {
	fmt.Printf("%s %s", m.say1, m.say2)
}

func main() {
	msg := Message{
		say1: "Hello",
		say2: "World",
	}

	msg.Print()
}
