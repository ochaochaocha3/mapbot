package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"

	"github.com/ochaochaocha3/mapbot/cmd/mapbot-repl/repl"
)

func main() {
	// 設定ファイルを読み込む
	configFile := "config.toml"
	config, err := repl.LoadConfigFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err)
		os.Exit(1)
	}

	// WindowsでもANSIエスケープシーケンスが正しく解釈されるように
	// colorable経由で標準出力を得る
	out := colorable.NewColorableStdout()

	// REPLを作り、起動する
	r := repl.New(os.Stdin, out, config)
	err = r.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "REPL error: %s\n", err)
		os.Exit(1)
	}
}
