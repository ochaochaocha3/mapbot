package repl

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/ochaochaocha3/mapbot/pkg/mapgen"
	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// REPL はREPLで使用するデータを格納する構造体。
type REPL struct {
	// in はREPLの入力源。
	in io.Reader
	// out はREPLの出力先。
	out io.Writer
	// terminated はREPLを終了するかを表す。
	terminated bool
	// completer は自動補完機能。
	completer *readline.PrefixCompleter

	// config はボットの設定。
	config *Config
	// fontCache はフォントデータの格納先。
	fontCache *mapgen.FontCache
	// squareMap はREPLセッション中に使用するスクエアマップ。
	squareMap *rpgmap.SquareMap
}

// New は新しいREPLを構築し、返す。
//
// REPLは、inから入力された文字列をコマンドとして実行し、
// outにその結果を出力する。
func New(in io.Reader, out io.Writer, config *Config) *REPL {
	completers := make([]readline.PrefixCompleterInterface, 0, len(commands))
	for _, c := range commands {
		completers = append(completers, readline.PcItem(c.Name, c.Completers...))
	}

	m, _ := rpgmap.NewSquareMap(10, 10)

	return &REPL{
		in:         in,
		out:        out,
		terminated: false,
		completer:  readline.NewPrefixCompleter(completers...),
		config:     config,
		squareMap:  m,
	}
}

// filterInput はreadlineでブロックする文字かどうかを判定する
func filterInput(r rune) (rune, bool) {
	switch r {
	// ^Z をブロックする
	// 現在は^Zを押すと動作がおかしくなるため
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// Start はREPLを開始する。
func (r *REPL) Start() error {
	// フォントを読み込む
	fc := mapgen.NewFontCache()
	err := fc.StoreFontDataFromFile(r.config.FontPath)
	if err != nil {
		return err
	}

	r.fontCache = fc

	// 自動補完機能を用意する
	l, err := readline.NewEx(&readline.Config{
		Prompt:              PROMPT,
		HistoryFile:         "mapbot-repl_history.txt",
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
		AutoComplete:        r.completer,
	})
	if err != nil {
		return err
	}
	defer l.Close()

	// チットの色の生成に乱数を使うため、乱数のシードを設定する
	rand.Seed(time.Now().UnixNano())

	// 動作開始
	r.printWelcomeMessage()

	// REPL
	for !r.terminated {
		line, readlineErr := l.Readline()

		switch readlineErr {
		case io.EOF:
			// ^D が押されたら修了する
			break
		case readline.ErrInterrupt:
			// ^C が押されたら次の読み込みに移る
			continue
		}

		line = strings.TrimSpace(line)

		// REPL終了の "q" のみ特別扱い
		if line == "q" {
			break
		}

		matches := commandRe.FindStringSubmatch(line)
		if matches == nil {
			r.printError(fmt.Errorf("無効なコマンドです: %s", line))
			continue
		}

		commandName := matches[1]
		cmd, ok := commandMap[commandName]
		if !ok {
			r.printError(fmt.Errorf("無効なコマンドです: %s", commandName))
			continue
		}

		args := tailSpacesRe.ReplaceAllString(matches[2], "")
		cmd.Handler(r, cmd, args)
	}

	return nil
}

// init はパッケージを初期化する。
func init() {
	registerCommands()
}
