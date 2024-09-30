package env

import "github.com/spf13/viper"

type Enviroment struct {
	ApiPort string `mapstructure:"PORT"`
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
