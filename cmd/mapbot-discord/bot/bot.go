package bot

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// ChannelToSquareMap はチャンネル -> マップの対応の型。
type ChannelToSquareMap map[string]*rpgmap.SquareMap

// Bot はマップ管理ボットの構造体。
type Bot struct {
	// Config はボットの設定。
	Config *Config
	// channelToMap はチャンネル -> マップの対応。
	channelToMap ChannelToSquareMap
	// mux は排他制御用のミューテックス。
	mux sync.Mutex
}

// NewBot は新しいボットを返す。
func NewBot(c *Config) *Bot {
	return &Bot{
		Config:       c,
		channelToMap: ChannelToSquareMap{},
	}
}

// Start はボットを起動する。
func (b *Bot) Start() error {
	// ボットの準備
	dg, err := discordgo.New("Bot " + b.Config.Token)
	if err != nil {
		return err
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		b.onMessageCreate(s, m)
	})

	// 通信開始
	err = dg.Open()
	if err != nil {
		return err
	}

	fmt.Println("Map Bot is now running. Press Ctrl-C to exit.")

	// シグナル処理
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()

	return nil
}

// onMessageCreate は発言時の処理。
//
// 発言に対応するコマンドがあれば実行する。
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		// 自分が発したメッセージは無視する
		return
	}

	input := strings.TrimSpace(m.Content)
	matches := commandRe.FindStringSubmatch(input)
	if matches == nil {
		return
	}

	commandName := matches[1]
	cmd, ok := commandMap[commandName]
	if !ok {
		return
	}

	commandArgs := strings.TrimSpace(matches[2])
	cmd.Handler(b, s, m, cmd, commandArgs)
}

// init はパッケージを初期化する。
func init() {
	initCommands()
}
