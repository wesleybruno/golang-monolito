package env

import "github.com/spf13/viper"

type Enviroment struct {
	ApiPort                 string `mapstructure:"PORT"`
	DbAddress               string `mapstructure:"DB_ADDRESS"`
	DbUser                  string `mapstructure:"DB_USER"`
	DbPassword              string `mapstructure:"DB_PASSWORD"`
	DbName                  string `mapstructure:"DB_NAME"`
	MaxOpenConns            int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns            int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	MaxIdleTime             string `mapstructure:"DB_MAX_IDLE_TIME"`
	Env                     string `mapstructure:"ENV"`
	ApiUrl                  string `mapstructure:"EXTERNAL_URL"`
	SendGridApiKey          string `mapstructure:"SENDGRID_API_KEY"`
	FromEmail               string `mapstructure:"FROM_EMAIL"`
	FrontendURL             string `mapstructure:"FRONTEND_URL"`
	AuthBasicUser           string `mapstructure:"AUTH_BASIC_USER"`
	AuthBasicPass           string `mapstructure:"AUTH_BASIC_PASS"`
	JwtSecret               string `mapstructure:"JWT_SECRET"`
	RedisAddr               string `mapstructure:"REDIS_ADDR"`
	RedisPwd                string `mapstructure:"REDIS_PWD"`
	RedisEnabled            bool   `mapstructure:"REDIS_ENABLED"`
	RateLimiterRequestCount int    `mapstructure:"RATE_LIMITER_REQUEST_COUNT"`
	RateLimiterEnabled      bool   `mapstructure:"RATE_LIMITER_ENABLED"`
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
