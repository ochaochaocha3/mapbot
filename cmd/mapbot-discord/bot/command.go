package bot

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/llgcode/draw2d/draw2dimg"

	"github.com/ochaochaocha3/mapbot/pkg/mapgen"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

const (
	COMMAND_INIT       = "init!"
	COMMAND_CLEAR      = "clear!"
	COMMAND_SIZE       = "size"
	COMMAND_LIST_CHITS = "lsc"
	COMMAND_ADD_CHIT   = "addc"
	COMMAND_MOVE_CHIT  = "mvc"
	COMMAND_HELP       = "help"

	// REPLY_MAP_NOT_FOUND はチャンネル用のマップが作成されていないことを表すメッセージ。
	REPLY_MAP_NOT_FOUND = "マップが作成されていません"
)

// コマンドハンドラの型。
type CommandHandler func(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	c *Command,
	argStr string,
)

// ボットのコマンドを表す構造体。
type Command struct {
	// コマンド名
	Name string
	// 引数の説明
	ArgsDescription string
	// 解説
	Description string
	// コマンドハンドラ
	Handler CommandHandler
}

// Usage はコマンドの使用方法の説明を返す。
func (c *Command) Usage() string {
	if c.ArgsDescription == "" {
		return fmt.Sprintf("`.%s`", c.Name)
	}

	return fmt.Sprintf("`.%s %s`", c.Name, c.ArgsDescription)
}

var (
	// commands は利用できるコマンド。
	commands []Command
	// commandMap はコマンド名とコマンドとの対応。
	commandMap = map[string]*Command{}

	// commandRe はコマンド実行を表す正規表現。
	commandRe = regexp.MustCompile(`\A\.([-!a-z]+)(?:\s+(.+))?`)
	// tailSpacesRe は末尾の空白を表す正規表現。
	tailSpacesRe = regexp.MustCompile(`\s+\z`)
)

func initCommands() {
	commands = []Command{
		{
			Name:            COMMAND_INIT,
			ArgsDescription: "幅 x 高さ",
			Description:     "マップを指定された大きさで初期化します（要注意！）",
			Handler:         initMap,
		},
		{
			Name:        COMMAND_CLEAR,
			Description: "マップを削除します（要注意！）",
			Handler:     clearMap,
		},
		{
			Name:        COMMAND_SIZE,
			Description: "マップの大きさを返します",
			Handler:     replyMapSize,
		},
		/*
			{
				Name:        COMMAND_LIST_CHITS,
				Description: "チットの一覧を出力します",
				Handler:     listChits,
			},
			{
				Name:            COMMAND_ADD_CHIT,
				ArgsDescription: `"チット名" (x, y)`,
				Description:     "チットを追加します",
				Handler:         addChit,
			},
			{
				Name:            COMMAND_MOVE_CHIT,
				ArgsDescription: `"チット名" (x, y)`,
				Description:     "チットを移動します",
				Handler:         moveChit,
			},
		*/
		{
			Name:        COMMAND_HELP,
			Description: "利用できるコマンドの使用法と説明を出力します",
			Handler:     replyHelp,
		},
	}

	for i, _ := range commands {
		c := &commands[i]
		commandMap[c.Name] = c
	}
}

// replyCommandUsage は、コマンドcの使用法を返信する。
func replyCommandUsage(c *Command, s *discordgo.Session, channelID string) {
	s.ChannelMessageSend(channelID, fmt.Sprintf("使用法: %s", c.Usage()))
}

// replyErrorMessage はエラー内容を返信する。
func replyErrorMessage(c *Command, err error, s *discordgo.Session, channelID string) {
	s.ChannelMessageSend(channelID, fmt.Sprintf(".%s: %s", c.Name, err))
}

// mapImageFilename はマップ画像のファイル名を返す。
func mapImageFilename(channelID string, c *Config) string {
	return filepath.Join(c.ImageDir, channelID+".png")
}

var initMapRe = regexp.MustCompile(`\A(\d+)\s*x\s*(\d+)\z`)

// initMap はマップを指定された大きさで初期化する。
func initMap(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	c *Command,
	argStr string,
) {
	matches := initMapRe.FindStringSubmatch(argStr)
	if matches == nil {
		replyCommandUsage(c, s, m.ChannelID)
		return
	}

	width, _ := strconv.Atoi(matches[1])
	height, _ := strconv.Atoi(matches[2])

	// クリティカルセクション：チャンネル用のマップを作って登録する
	b.Lock()

	newMap, err := rpgmap.NewSquareMap(width, height)
	if err == nil {
		b.channelToMap[m.ChannelID] = newMap
	}

	// クリティカルセクション終了
	b.Unlock()

	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	// マップの画像を作る
	mImg := mapgen.NewSquareMapImage(newMap)
	i, err := mImg.Render()
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	// マップの画像を保存する
	filename := mapImageFilename(m.ChannelID, b.Config)
	err = draw2dimg.SaveToPngFile(filename, i)
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	// マップの画像を読み込み、メッセージに添付して送信する
	f, err := os.Open(filename)
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}
	defer f.Close()

	msgData := discordgo.MessageSend{
		Content: newMap.String(),
		File: &discordgo.File{
			Name:        filename,
			ContentType: "image/png",
			Reader:      f,
		},
	}

	s.ChannelMessageSendComplex(m.ChannelID, &msgData)
}

// clearMap はマップを削除する。
func clearMap(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	_ *Command,
	_ string,
) {
	// クリティカルセクション：チャンネル用のマップを削除する
	b.Lock()

	_, found := b.channelToMap[m.ChannelID]
	if found {
		delete(b.channelToMap, m.ChannelID)
		os.Remove(mapImageFilename(m.ChannelID, b.Config))
	}

	// クリティカルセクション終了
	b.Unlock()

	if !found {
		s.ChannelMessageSend(m.ChannelID, REPLY_MAP_NOT_FOUND)
		return
	}

	s.ChannelMessageSend(m.ChannelID, "マップを削除しました")
}

// replyMapSize はマップの大きさを返信する。
func replyMapSize(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	_ *Command,
	_ string,
) {
	sMap, found := b.channelToMap[m.ChannelID]
	if !found {
		s.ChannelMessageSend(m.ChannelID, REPLY_MAP_NOT_FOUND)
		return
	}

	s.ChannelMessageSend(m.ChannelID, sMap.SizeStr())
}

// replyHelp は、利用できるコマンドの使用法と説明を返信する。
func replyHelp(
	_ *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	_ *Command,
	_ string,
) {
	var buf bytes.Buffer
	for _, c := range commands {
		buf.WriteString("`")
		buf.WriteString(".")
		buf.WriteString(c.Name)
		if c.ArgsDescription != "" {
			buf.WriteString(" ")
			buf.WriteString(c.ArgsDescription)
		}
		buf.WriteString("`\n    ")
		buf.WriteString(c.Description)
		buf.WriteString("\n")
	}

	s.ChannelMessageSend(m.ChannelID, strings.TrimSpace(buf.String()))
}
