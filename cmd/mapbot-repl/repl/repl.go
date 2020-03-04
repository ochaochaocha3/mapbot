package repl

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

// REPL はREPLで使用するデータを格納する構造体。
type REPL struct {
	in         io.Reader
	out        io.Writer
	terminated bool
	completer  *readline.PrefixCompleter

	squareMap *rpgmap.SquareMap
}

// New は新しいREPLを構築し、返す。
//
// REPLは、inから入力された文字列をコマンドとして実行し、
// outにその結果を出力する。
func New(in io.Reader, out io.Writer) *REPL {
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
func (r *REPL) Start() {
	l, err := readline.NewEx(&readline.Config{
		Prompt:              PROMPT,
		HistoryFile:         "mapbot-repl_history.txt",
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
		AutoComplete:        r.completer,
	})
	if err != nil {
		r.printError(err)
		return
	}
	defer l.Close()

	rand.Seed(time.Now().UnixNano())

	r.printWelcomeMessage()

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

		commandArgs := tailSpacesRe.ReplaceAllString(matches[2], "")
		cmd.Handler(r, cmd, commandArgs)
	}
}

// init はパッケージを初期化する。
func init() {
	registerCommands()
}
