package config

import (
	"encoding/json"
	"flag"
	"github.com/pkg/errors"
	"io/ioutil"
)

type BaseConfig struct {
	ServiceName string `json:"service_name"`
	ServerID    string `json:"server_id"`
}

type DatabaseSettings struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Database        string `json:"database"`
	Password        string `json:"password"`
	Timeout         int    `json:"timeout"`
	PoolSize        int    `json:"pool_size"`
	RetriesNum      int    `json:"retries_num"`
	RetriesInterval int    `json:"retries_interval"`
}

type MongoDBSettings struct {
	DatabaseSettings `json:",inline"`
	Collection       string `json:"collection"`
}

type QueueSettings struct {
	FetchDelay int `json:"fetch_delay"`
	WorkersNum int `json:"workers_num"`
}

type MongoQueue struct {
	MongoDBSettings `json:",inline"`
	QueueSettings   `json:",inline"`
}

type TelegramSettings struct {
	APIToken    string `json:"api_token"`
	BotName     string `json:"bot_name"`
	HttpTimeout int    `json:"http_timeout"`
	Retries     int    `json:"retries"`
}

type TelegramPolling struct {
	PollTimeout int `json:"poll_timeout"`
	RetryDelay  int `json:"retry_delay"`
}

type Logging struct {
	DefaultLevel string `json:"default_level"`
	TogglePort   int    `json:"toggle_port"`
	TogglePath   string `json:"toggle_path"`
}

type GoogleAPI struct {
	APIKey      string `json:"api_key"`
	HttpTimeout int    `json:"http_timeout"`
}

func FromJSONFile(path string, config interface{}) error {
	if path == "" {
		return errors.New("json config: empty path")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "json config: cannot read file")
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return errors.Wrap(err, "json config: cannot parse file content")
	}
	return nil
}

func FromCmdArgs(confPath *string) {
	flag.StringVar(confPath, "conf", "conf.json", "Path to the config file")
}
