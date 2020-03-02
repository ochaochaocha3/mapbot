package repl

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/chzyer/readline"

	"github.com/ochaochaocha3/mapbot/pkg/rpgmap"
)

const (
	// 書式設定をリセットするエスケープシーケンス
	ESC_RESET = "\033[0m"
	// 太字にするエスケープシーケンス
	ESC_BOLD = "\033[1m"
	// 文字色を赤色にするエスケープシーケンス
	ESC_RED = "\033[31m"
	// 文字色を黄色にするエスケープシーケンス
	ESC_YELLOW = "\033[33m"
	// 文字色をシアンにするエスケープシーケンス
	ESC_CYAN = "\033[36m"

	// REPLのプロンプト
	PROMPT = ESC_YELLOW + ">>" + ESC_RESET + " "
	// 結果の初めに出力する文字列
	RESULT_HEADER = ESC_CYAN + "=>" + ESC_RESET + " "

	COMMAND_INIT       = "init"
	COMMAND_SIZE       = "size"
	COMMAND_LIST_CHITS = "lsc"
	COMMAND_ADD_CHIT   = "addc"
	COMMAND_MOVE_CHIT  = "mvc"
	COMMAND_HELP       = "help"
	COMMAND_QUIT       = "quit"
)

// コマンドハンドラの型。
type CommandHandler func(r *REPL, c *Command, input string)

// REPLコマンドを表す構造体。
type Command struct {
	// コマンド名
	Name string
	// 引数の説明
	ArgsDescription string
	// 解説
	Description string
	// コマンドハンドラ
	Handler CommandHandler
	// 自動補完の候補
	Completers []readline.PrefixCompleterInterface
}

// Usage はコマンドの使用方法の説明を返す。
func (c *Command) Usage() string {
	if c.ArgsDescription == "" {
		return c.Name
	}

	return c.Name + " " + c.ArgsDescription
}

var (
	// commands は利用できるコマンド。
	commands []Command
	// commandMap はコマンド名とコマンドとの対応。
	commandMap = map[string]*Command{}

	// commandRe はコマンド実行を表す正規表現。
	commandRe = regexp.MustCompile(`\A([-a-z]+)(?:\s+(.+))*`)
	// tailSpacesRe は末尾の空白を表す正規表現。
	tailSpacesRe = regexp.MustCompile(`\s+\z`)
)

// REPL はREPLで使用するデータを格納する構造体。
type REPL struct {
	in         io.Reader
	out        io.Writer
	terminated bool
	completer  *readline.PrefixCompleter

	squareMap *rpgmap.SquareMap
}

// init はパッケージを初期化する。
func init() {
	commands = []Command{
		{
			Name:            COMMAND_INIT,
			ArgsDescription: "幅 x 高さ",
			Description:     "マップを指定された大きさで初期化します",
			Handler:         initMap,
		},
		{
			Name:        COMMAND_SIZE,
			Description: "マップの大きさを出力します",
			Handler:     printSize,
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
			Name:            COMMAND_MOVE_CHIT,
			ArgsDescription: `"チット名" (x, y)`,
			Description:     "チットを移動します",
			Handler:         moveChit,
		},
		{
			Name:        COMMAND_HELP,
			Description: "利用できるコマンドの使用法と説明を出力します",
			Handler:     printHelp,
		},
		{
			Name:        COMMAND_QUIT,
			Description: "mapbot REPLを終了します",
			Handler:     terminateREPL,
		},
	}

	for i, _ := range commands {
		c := &commands[i]
		commandMap[c.Name] = c
	}
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

// printCommandUsage は、コマンドcの使用法を出力する。
func (r *REPL) printCommandUsage(c *Command) {
	fmt.Fprintf(r.out, "使用法: %s\n", c.Usage())
}

// printWelcomeMessage は起動時の歓迎メッセージを出力する。
func (r *REPL) printWelcomeMessage() {
	fmt.Fprintln(r.out, ESC_BOLD+"mapbot REPL"+ESC_RESET)
	fmt.Fprintln(r.out, "\n* \"help\" と入力すると、利用できるコマンドの使用法と説明を出力します")
	fmt.Fprintln(r.out, "* \"q\" または \"quit\" と入力すると終了します")
	fmt.Fprintln(r.out, "")
}

// printOK はコマンドの実行に成功した旨のメッセージを出力する。
func (r *REPL) printOK() {
	fmt.Fprintln(r.out, ESC_CYAN+"OK"+ESC_RESET)
}

// printError はエラーメッセージを強調して出力する。
func (r *REPL) printError(err error) {
	fmt.Fprintln(r.out, ESC_RED+err.Error()+ESC_RESET)
}

var initMapRe = regexp.MustCompile(`\A(\d+)\s*x\s*(\d+)\z`)

// initMap はマップを指定された大きさで初期化する。
func initMap(r *REPL, c *Command, input string) {
	m := initMapRe.FindStringSubmatch(input)
	if m == nil {
		r.printCommandUsage(c)
		return
	}

	width, _ := strconv.Atoi(m[1])
	height, _ := strconv.Atoi(m[2])

	newMap, err := rpgmap.NewSquareMap(width, height)
	if err != nil {
		r.printError(err)
		return
	}

	r.squareMap = newMap

	r.printOK()
}

// printSize はマップの大きさを出力する。
func printSize(r *REPL, _ *Command, _ string) {
	fmt.Fprintf(r.out, "%s%s\n", RESULT_HEADER, r.squareMap.SizeStr())
}

// listChits はチットの一覧を出力する。
func listChits(r *REPL, _ *Command, _ string) {
	for _, c := range r.squareMap.Chits() {
		fmt.Fprintln(r.out, c)
	}
}

var addChitRe = regexp.MustCompile(`\A"([^"]+)"\s*\((\d+),\s*(\d+)\)\z`)

// addChit はチットを追加する。
func addChit(r *REPL, c *Command, input string) {
	m := addChitRe.FindStringSubmatch(input)
	if m == nil {
		r.printCommandUsage(c)
		return
	}

	name := m[1]
	x, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])

	chit := rpgmap.Chit{
		Name: name,
		X:    x - 1,
		Y:    y - 1,
	}

	err := r.squareMap.AddChit(&chit)
	if err != nil {
		r.printError(err)
		return
	}

	fmt.Fprintf(r.out, "%s%s\n", RESULT_HEADER, chit.String())
}

// moveChit はチットを移動する。
func moveChit(r *REPL, c *Command, input string) {
	m := addChitRe.FindStringSubmatch(input)
	if m == nil {
		r.printCommandUsage(c)
		return
	}

	name := m[1]
	x, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])

	err := r.squareMap.MoveChit(name, x-1, y-1)
	if err != nil {
		r.printError(err)
		return
	}

	r.printOK()
}

// printHelp は、利用できるコマンドの使用法と説明を出力する。
func printHelp(r *REPL, _ *Command, _ string) {
	for _, c := range commands {
		fmt.Fprint(r.out, ESC_BOLD+c.Name+ESC_RESET)
		if c.ArgsDescription != "" {
			fmt.Fprint(r.out, " "+c.ArgsDescription)
		}
		fmt.Fprintln(r.out, "")

		fmt.Fprintln(r.out, "    "+c.Description)
	}
}

// terminateREPL はREPLを終了させる。
func terminateREPL(r *REPL, _ *Command, _ string) {
	r.terminated = true
}
