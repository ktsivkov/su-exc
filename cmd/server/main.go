package main

import (
	"context"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	"github.com/ktsivkov/su-exc/internal/rest"
)

func main() {
	conf := viper.New()
	conf.AddConfigPath("configs")
	conf.SetConfigType("yaml")
	conf.SetConfigName("app")
	conf.AutomaticEnv()
	if err := conf.ReadInConfig(); err != nil {
		panic(err)
	}

	dbUri := conf.GetString("POSTGRES_URI")
	appPort := conf.GetInt("APP_PORT")
	shutdownGracePeriod := conf.GetDuration("APP_SHUTDOWN_GRACE_PERIOD")

	err := rest.Boot(context.Background(), dbUri, appPort, shutdownGracePeriod)
	if err != nil {
		panic(err)
	}
}
