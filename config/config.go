package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"os"
	"strconv"
)

// 定义全局的配置变量
var (
	confPath  string
	deployEnv string
	host      string
	addrs     string
	debug     bool
	Conf      *Config // config
)

func init() {
	var (
		defHost, _  = os.Hostname()
		defAddrs    = os.Getenv("ADDRS")
		defDebug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	)
	flag.StringVar(&confPath, "conf", "./config/config.toml", "default config path.")
	flag.StringVar(&deployEnv, "deploy.env", os.Getenv("DEPLOY_ENV"), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	flag.StringVar(&host, "host", defHost, "machine hostname. or use default machine hostname.")
	flag.StringVar(&addrs, "addrs", defAddrs, "server public ip addrs. or use ADDRS env variable, value: 127.0.0.1 etc.")
	flag.BoolVar(&debug, "debug", defDebug, "server debug. or use DEBUG env variable, value: true/false etc.")
}

// Init init config 初始化配置文件中的参数.
func Init() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Config is comet config.
type Config struct {
	APPID uint64 `toml:"appid"`
	TOKEN string `toml:"token"`
	Debug bool   `toml:"debug"`
}
