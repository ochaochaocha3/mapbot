package bot

import (
	"github.com/BurntSushi/toml"
)

// Config はボットの設定の構造体。
type Config struct {
	// Token はボットアカウントのトークン。
	Token string
	// ImageDir は画像を格納するディレクトリ。
	ImageDir string
}

// LoadConfigFile は設定ファイルを読み込み、Config構造体を返す。
func LoadConfigFile(filename string) (*Config, error) {
	config := Config{}

	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return nil, err
	}

	if config.ImageDir == "" {
		config.ImageDir = "."
	}

	return &config, nil
}
