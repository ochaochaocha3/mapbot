package bot

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ochaochaocha3/mapbot/pkg/mapgen"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// ChannelToSquareMap はチャンネル -> マップの対応の型。
type ChannelToSquareMap map[string]*rpgmap.SquareMap

// Bot はマップ管理ボットの構造体。
type Bot struct {
	// config はボットの設定。
	config *Config
	// fontCache はフォントデータの格納先。
	fontCache *mapgen.FontCache
	// channelToMap はチャンネル -> マップの対応。
	channelToMap ChannelToSquareMap
	// mux は排他制御用のミューテックス。
	mux sync.Mutex
}

// New は新しいボットを返す。
func New(c *Config) *Bot {
	return &Bot{
		config:       c,
		channelToMap: ChannelToSquareMap{},
	}
}

// Start はボットを起動する。
func (b *Bot) Start() error {
	// フォントを読み込む
	fc := mapgen.NewFontCache()
	err := fc.StoreFontDataFromFile(b.config.FontPath)
	if err != nil {
		return err
	}

	b.fontCache = fc

	// ボットを準備する
	dg, err := discordgo.New("Bot " + b.config.Token)
	if err != nil {
		return err
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		b.onMessageCreate(s, m)
	})

	// チットの色の生成に乱数を使うため、乱数のシードを設定する
	rand.Seed(time.Now().UnixNano())

	// 通信開始
	err = dg.Open()
	if err != nil {
		return err
	}
	defer dg.Close()

	fmt.Println("Map Bot is now running. Press Ctrl-C to exit.")

	// シグナル処理
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

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

	args := strings.TrimSpace(matches[2])
	cmd.Handler(b, s, m, cmd, args)
}

// init はパッケージを初期化する。
func init() {
	registerCommands()
}
