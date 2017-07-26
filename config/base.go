package config

import (
	"encoding/json"
	"flag"
	"github.com/gazoon/bot_libs/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
)

type BaseConfig struct {
	ServiceName string `mapstructure:"service_name" json:"service_name"`
	ServerID    string `mapstructure:"server_id" json:"server_id"`
	Port        int    `mapstructure:"port" json:"port"`
}

type DatabaseSettings struct {
	Host            string `mapstructure:"host" json:"host"`
	Port            int    `mapstructure:"port" json:"port"`
	User            string `mapstructure:"user" json:"user"`
	Database        string `mapstructure:"database" json:"database"`
	Password        string `mapstructure:"password" json:"password"`
	Timeout         int    `mapstructure:"timeout" json:"timeout"`
	PoolSize        int    `mapstructure:"pool_size" json:"pool_size"`
	RetriesNum      int    `mapstructure:"retries_num" json:"retries_num"`
	RetriesInterval int    `mapstructure:"retries_interval" json:"retries_interval"`
}

type S3Setting struct {
	Region string `mapstructure:"region" json:"region"`
	Bucket string `mapstructure:"bucket" json:"bucket"`
}

type AwsCreds struct {
	AccountID     string `mapstructure:"account_id" json:"account_id"`
	AccountSecret string `mapstructure:"account_secret" json:"account_secret"`
}

type MongoDBSettings struct {
	DatabaseSettings `mapstructure:",squash" json:",inline"`
	Collection       string `mapstructure:"collection" json:"collection"`
}

type QueueSettings struct {
	FetchDelay int `mapstructure:"fetch_delay" json:"fetch_delay"`
	WorkersNum int `mapstructure:"workers_num" json:"workers_num"`
}

type MongoQueue struct {
	MongoDBSettings `mapstructure:",squash" json:",inline"`
	QueueSettings   `mapstructure:",squash" json:",inline"`
}

type TelegramSettings struct {
	APIToken    string `mapstructure:"api_token" json:"api_token"`
	BotName     string `mapstructure:"bot_name" json:"bot_name"`
	HttpTimeout int    `mapstructure:"http_timeout" json:"http_timeout"`
	Retries     int    `mapstructure:"retries" json:"retries"`
}

type TelegramPolling struct {
	PollTimeout int `mapstructure:"poll_timeout" json:"poll_timeout"`
	RetryDelay  int `mapstructure:"retry_delay" json:"retry_delay"`
}

type Logging struct {
	DefaultLevel string `mapstructure:"default_level" json:"default_level"`
	TogglePort   int    `mapstructure:"toggle_port" json:"toggle_port"`
	TogglePath   string `mapstructure:"toggle_path" json:"toggle_path"`
}

type GoogleAPI struct {
	APIKey      string `mapstructure:"api_key" json:"api_key"`
	HttpTimeout int    `mapstructure:"http_timeout" json:"http_timeout"`
}

func FromJSONFile(path string, config interface{}) error {
	err := initializeConfig(path, config, json.Unmarshal)
	return errors.Wrap(err, "json config")
}

func initializeConfig(path string, config interface{}, parser func([]byte, interface{}) error) error {
	data, err := getFileData(path, parser)
	if err != nil {
		return errors.Wrap(err, "original file")
	}
	var localData map[string]interface{}
	pathToLocalConf := buildLocalPath(path)
	if _, err := os.Stat(pathToLocalConf); !os.IsNotExist(err) {
		localData, err = getFileData(pathToLocalConf, parser)
		if err != nil {
			return errors.Wrap(err, "local file")
		}
	}
	resultData := utils.MergeMaps(data, localData)
	err = mapstructure.Decode(resultData, config)
	if err != nil {
		return errors.Wrap(err, "converting to the config structure failed")
	}
	return nil
}

func getFileData(path string, parser func([]byte, interface{}) error) (map[string]interface{}, error) {
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read file")
	}
	data := map[string]interface{}{}
	err = parser(fileContent, &data)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse file content")
	}
	return data, nil
}

func buildLocalPath(filePath string) string {
	fileName := path.Base(filePath)
	localName := "local_" + fileName
	if fileName == filePath {
		return localName
	}
	return path.Dir(filePath) + localName
}

func FromCmdArgs(confPath *string) {
	flag.StringVar(confPath, "conf", "conf.json", "Path to the config file")
}
