package configs

import (
	"github.com/spf13/viper"
)

var Cfg *conf

type conf struct {
	DBDriver   string `mapstructure:"DB_DRIVER"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBUrl      string `mapstructure:"DB_URL"`

	RdHost     string `mapstructure:"RD_HOST"`
	RdPort     string `mapstructure:"RD_PORT"`
	RdPassword string `mapstructure:"RD_PASSWORD"`

	WebServerHost string `mapstructure:"WEB_SERVER_HOST"`
	WebServerPort string `mapstructure:"WEB_SERVER_PORT"`
	WebServerUrl  string `mapstructure:"WEB_SERVER_URL"`

	PathUploads string `mapstructure:"PATH_UPLOADS"`
	UrlUploads  string `mapstructure:"URL_UPLOADS"`

	GinMode string `mapstructure:"GIN_MODE"`

	JWTSecretKey  string `mapstructure:"JWT_SECRET_KEY"`
	JWTExpriresIn string `mapstructure:"JWT_EXPIRES_IN"`
}

func LoadConfig(path string) *conf {
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		panic(err)
	}

	return Cfg
}

func GetConfig() *conf {
	return Cfg
}
