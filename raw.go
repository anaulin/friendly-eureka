package main

import (
	"flag"
	"net/url"
	"sort"
	"strconv"
	"time"

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

	// Get credentials from env variables.
	var c Config
	err := envconfig.Process("raw", &c)
	if err != nil {
		glog.Fatal(err.Error())
	}

	// Prep our API object.
	anaconda.SetConsumerKey(c.TwitConsumerKey)
	anaconda.SetConsumerSecret(c.TwitConsumerSeekret)
	api := anaconda.NewTwitterApi(c.TwitAccessToken, c.TwitAccessTokenSeekret)

	// Get self User. We'll use this to know our own id.
	self, err := api.GetSelf(nil)
	if err != nil {
		glog.Fatal("Couldnt get self: ", err)
	}

	// Ad infinitum...
	for {
		v := url.Values{}
		v.Set("result_type", "recent")
		v.Set("count", "100")
		v.Set("since_id", strconv.FormatInt(getLastTweetId(api, self.Id), 10))

		searchAndRetweet("ridiculous filter:media", api, &v)
		searchAndRetweet("whimsical filter:media", api, &v)

		time.Sleep(1 * time.Minute)
	}

	glog.Flush()
}

// Implementation of sort.Interface for []anaconda.Tweet
// Sorting based on retweet and favorite count.
type ByFavAndRetweet []anaconda.Tweet

func (tweets ByFavAndRetweet) Len() int {
	return len(tweets)
}

func (tweets ByFavAndRetweet) Swap(i, j int) {
	tweets[i], tweets[j] = tweets[j], tweets[i]
}

func (tweets ByFavAndRetweet) Less(i, j int) bool {
	return ((tweets[i].RetweetCount + tweets[i].FavoriteCount) <
		(tweets[j].RetweetCount + tweets[j].FavoriteCount))
}

// End sort.Interface implementation

// Returns the id of our most recent Tweet.
func getLastTweetId(api *anaconda.TwitterApi, id int64) int64 {
	v := url.Values{}
	v.Set("user_id", strconv.FormatInt(id, 10))
	v.Set("trim_user", "1") // no need to return full user object
	v.Set("exclude_replies", "1")
	v.Set("include_rts", "1")

	tweets, err := api.GetUserTimeline(v)
	if err != nil {
		glog.Error("Couldnt get own timeline: ", err)
		return 1
	}
	if len(tweets) < 1 {
		return 0
	}

	return tweets[0].Id
}

func searchAndRetweet(query string, api *anaconda.TwitterApi, v *url.Values) {
	glog.Infof("Searching with params: %+v", *v)
	result, err := api.GetSearch(query, *v)
	if err != nil {
		glog.Error("Error searching: ", err)
	}
	if len(result.Statuses) == 0 {
		glog.Warning("Search returned zero results")
		return
	}

	sort.Sort(ByFavAndRetweet(result.Statuses))
	// Often the retweet fails because it has already been retweeted by us.
	// Iterate through the list until something gets successfully retweeted.
	retweeted := false
	i := len(result.Statuses) - 1
	for !retweeted && i >= 0 {
		_, err = api.Retweet(result.Statuses[i].Id, true)
		i = i - 1
		if err != nil {
			glog.Error("Error retweeting: ", err)
		} else {
			retweeted = true
		}
	}
}
