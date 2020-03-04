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

	"github.com/ochaochaocha3/mapbot/pkg/colorutil"
	"github.com/ochaochaocha3/mapbot/pkg/mapgen"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

const (
	COMMAND_INIT        = "init!"
	COMMAND_CLEAR       = "clear!"
	COMMAND_SIZE        = "size"
	COMMAND_LIST_CHITS  = "lsc"
	COMMAND_ADD_CHIT    = "addc"
	COMMAND_DELETE_CHIT = "delc"
	COMMAND_MOVE_CHIT   = "mvc"
	COMMAND_HELP        = "help"

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

// registerCommands はボットのコマンドを登録する。
func registerCommands() {
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
			Name:            COMMAND_DELETE_CHIT,
			ArgsDescription: `"チット名"`,
			Description:     "チットを削除します",
			Handler:         deleteChit,
		},
		{
			Name:            COMMAND_MOVE_CHIT,
			ArgsDescription: `"チット名" (x, y)`,
			Description:     "チットを移動します",
			Handler:         moveChit,
		},
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
	b.mux.Lock()

	newMap, err := rpgmap.NewSquareMap(width, height)
	if err == nil {
		b.channelToMap[m.ChannelID] = newMap
	}

	// クリティカルセクション終了
	b.mux.Unlock()

	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	err = uploadMap(&UploadMapArgs{
		Content:   newMap.String(),
		Map:       newMap,
		Session:   s,
		ChannelID: m.ChannelID,
		Config:    b.Config,
	})
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
	}
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
	b.mux.Lock()

	_, found := b.channelToMap[m.ChannelID]
	if found {
		delete(b.channelToMap, m.ChannelID)
		os.Remove(mapImageFilename(m.ChannelID, b.Config))
	}

	// クリティカルセクション終了
	b.mux.Unlock()

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

// listChits はチットの一覧を出力する。
func listChits(
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

	if sMap.NumOfChits() < 1 {
		s.ChannelMessageSend(m.ChannelID, "（チット未登録）")
		return
	}

	chitStrs := make([]string, 0, sMap.NumOfChits())
	sMap.ForEachChit(func(_ int, c *rpgmap.Chit) {
		chitStrs = append(chitStrs, c.String())
	})

	s.ChannelMessageSend(m.ChannelID, strings.Join(chitStrs, "\n"))
}

var addChitRe = regexp.MustCompile(`\A"([^"]+)"\s*\((\d+),\s*(\d+)\)\z`)

// addChit はチットを追加する。
func addChit(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	c *Command,
	argStr string,
) {
	matches := addChitRe.FindStringSubmatch(argStr)
	if matches == nil {
		replyCommandUsage(c, s, m.ChannelID)
		return
	}

	sMap, found := b.channelToMap[m.ChannelID]
	if !found {
		s.ChannelMessageSend(m.ChannelID, REPLY_MAP_NOT_FOUND)
		return
	}

	name := matches[1]
	x, _ := strconv.Atoi(matches[2])
	y, _ := strconv.Atoi(matches[3])

	chit := rpgmap.Chit{
		Name:  name,
		X:     x - 1,
		Y:     y - 1,
		Color: colorutil.RandomChitColor(),
	}

	err := sMap.AddChit(&chit)
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	err = uploadMap(&UploadMapArgs{
		Content:   chit.String(),
		Map:       sMap,
		Session:   s,
		ChannelID: m.ChannelID,
		Config:    b.Config,
	})
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
	}
}

var deleteChitRe = regexp.MustCompile(`\A"([^"]+)"\z`)

// deleteChit はチットを削除する。
func deleteChit(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	c *Command,
	argStr string,
) {
	matches := deleteChitRe.FindStringSubmatch(argStr)
	if matches == nil {
		replyCommandUsage(c, s, m.ChannelID)
		return
	}

	sMap, found := b.channelToMap[m.ChannelID]
	if !found {
		s.ChannelMessageSend(m.ChannelID, REPLY_MAP_NOT_FOUND)
		return
	}

	name := matches[1]

	err := sMap.DeleteChit(name)
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	err = uploadMap(&UploadMapArgs{
		Content:   fmt.Sprintf("チット「%s」を削除しました", name),
		Map:       sMap,
		Session:   s,
		ChannelID: m.ChannelID,
		Config:    b.Config,
	})
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
	}
}

// moveChit はチットを移動する。
func moveChit(
	b *Bot,
	s *discordgo.Session,
	m *discordgo.MessageCreate,
	c *Command,
	argStr string,
) {
	matches := addChitRe.FindStringSubmatch(argStr)
	if matches == nil {
		replyCommandUsage(c, s, m.ChannelID)
		return
	}

	sMap, found := b.channelToMap[m.ChannelID]
	if !found {
		s.ChannelMessageSend(m.ChannelID, REPLY_MAP_NOT_FOUND)
		return
	}

	name := matches[1]
	x, _ := strconv.Atoi(matches[2])
	y, _ := strconv.Atoi(matches[3])

	chit, err := sMap.MoveChit(name, x-1, y-1)
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
		return
	}

	err = uploadMap(&UploadMapArgs{
		Content:   chit.String(),
		Map:       sMap,
		Session:   s,
		ChannelID: m.ChannelID,
		Config:    b.Config,
	})
	if err != nil {
		replyErrorMessage(c, err, s, m.ChannelID)
	}
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

// UploadMapArgs はマップアップロードに必要な情報の構造体。
type UploadMapArgs struct {
	// Content は画像とともに送信する文字列。
	Content string
	// Map はスクエアマップ。
	Map *rpgmap.SquareMap
	// Session はDiscordボットのセッション。
	Session *discordgo.Session
	// ChannelID はチャンネルのID。
	ChannelID string
	// Config はボットの設定。
	Config *Config
}

// uploadMap はマップを描画してアップロードする。
func uploadMap(args *UploadMapArgs) error {
	// マップの画像を作る
	mImg := mapgen.NewSquareMapImage(args.Map)
	i, err := mImg.Render()
	if err != nil {
		return err
	}

	// マップの画像を保存する
	filename := mapImageFilename(args.ChannelID, args.Config)
	err = draw2dimg.SaveToPngFile(filename, i)
	if err != nil {
		return err
	}

	// マップの画像を読み込み、メッセージに添付して送信する
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	msgData := discordgo.MessageSend{
		Content: args.Content,
		File: &discordgo.File{
			Name:        filename,
			ContentType: "image/png",
			Reader:      f,
		},
	}

	args.Session.ChannelMessageSendComplex(args.ChannelID, &msgData)

	return nil
}
