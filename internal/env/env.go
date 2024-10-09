package env

import "github.com/spf13/viper"

type Enviroment struct {
	ApiPort      string `mapstructure:"PORT"`
	DbAddress    string `mapstructure:"DB_ADDRESS"`
	DbUser       string `mapstructure:"DB_USER"`
	DbPassword   string `mapstructure:"DB_PASSWORD"`
	DbName       string `mapstructure:"DB_NAME"`
	MaxOpenConns int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MaxIdleTime  string `mapstructure:"DB_MAX_IDLE_TIME"`
	Env          string `mapstructure:"ENV"`
}

var Config Enviroment

func LoadConfig() (*Enviroment, error) {
	var cfg *Enviroment

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	Config = *cfg
	return cfg, err
}
