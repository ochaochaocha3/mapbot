package main

import (
	"github.com/mattn/go-colorable"
	"github.com/ochaochaocha3/mapbot/cmd/mapbot-repl/repl"
	"os"
)

func main() {
	// WindowsでもANSIエスケープシーケンスが正しく解釈されるように
	// colorable経由で標準出力を得る
	out := colorable.NewColorableStdout()

	r := repl.New(os.Stdin, out)
	r.Start()
}
