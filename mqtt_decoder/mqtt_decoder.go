package main

import (
	"fmt"
	mqtt "github.com/huin/mqtt"
	"io"
	"os"
)

func FprintMessage(f io.Writer, msg mqtt.Message) {
  fmt.Fprintf(f, "%#v", msg)
}

func FprintError(f io.Writer, err error) {
  if err != nil {
    fmt.Fprintf(f, " // %s\n", err)
  }
}

func main() {
	input := os.Stdin
  output := os.Stdout

	for {
		msg, err := mqtt.DecodeOneMessage(input, nil)
    if err == io.EOF {
      os.Exit(0)
    }
    FprintMessage(output, msg)
    FprintError(output, err)
	}
}
