package repl

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config はREPLの設定の構造体。
type Config struct {
	// ImageDir は画像を格納するディレクトリ。
	ImageDir string
	// FontPath はTrueTypeフォントファイルのパス。
	FontPath string
}

// LoadConfigFile は設定ファイルを読み込み、Config構造体を返す。
func LoadConfigFile(filename string) (*Config, error) {
	config := Config{}

	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return nil, err
	}

	if config.FontPath == "" {
		return nil, fmt.Errorf("FontPath is not set")
	}

	if config.ImageDir == "" {
		config.ImageDir = "."
	}

	return &config, nil
}
