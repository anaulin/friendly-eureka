package main

import (
	"flag"

	"github.com/ChimeraCoder/anaconda"
	"github.com/golang/glog"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TwitConsumerKey        string `envconfig:"twit_consumer_key"`
	TwitConsumerSeekret    string `envconfig:"twit_consumer_seekret"`
	TwitAccessToken        string `envconfig:"twit_access_token"`
	TwitAccessTokenSeekret string `envconfig:"twit_access_token_seekret"`
}

func main() {
	flag.Parse() // needed for glog

	var c Config
	err := envconfig.Process("raw", &c)
	if err != nil {
		glog.Fatal(err.Error())
	}
	glog.Infof("Config: %+v", c)

	println("hello world")

	anaconda.SetConsumerKey(c.TwitConsumerKey)
	anaconda.SetConsumerSecret(c.TwitConsumerSeekret)
	api := anaconda.NewTwitterApi(c.TwitAccessToken, c.TwitAccessTokenSeekret)

	searchResult, err := api.GetSearch("ridiculous", nil)
	if err != nil {
		glog.Error("Error searching: ", err)
	}
	for _, tweet := range searchResult.Statuses {
		println(tweet.Text)
	}
	glog.Flush()
}
