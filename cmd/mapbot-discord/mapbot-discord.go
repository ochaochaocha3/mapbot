package main

import (
	"fmt"
	"os"

	"github.com/ochaochaocha3/mapbot/cmd/mapbot-discord/bot"
)

func main() {
	// 設定ファイルを読み込む
	configFile := "config.toml"
	config, err := bot.LoadConfigFrom(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err)
		os.Exit(1)
	}

	// ボットを作り、起動する
	b := bot.NewBot(config)

	err = b.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bot error: %s\n", err)
		os.Exit(1)
	}
}
