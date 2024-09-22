package models

import (
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

type Dimension struct {
	APPID   string
	Country string
	OS      string
}

type Response struct {
	CampaignID string `bson:"campaign_id" json:"cid"`
	Image      string `bson:"image" json:"img"`
	CTA        string `bson:"cta" json:"cta"`
}

type Helpers struct {
	AppName string
	AppPort string
	DB      *mongo.Database
	Logger  *slog.Logger
	Redis   *redis.Client
	Metrics *Metrics
}

type Metrics struct {
	RequestCounter  *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorCounter    *prometheus.CounterVec
	CacheHits       *prometheus.CounterVec
	CacheMisses     *prometheus.CounterVec
}

type Rule struct {
	Dimension string   `bson:"dimension" json:"dimension"`
	Include   []string `bson:"include" json:"include"`
	Exclude   []string `bson:"exclude" json:"exclude"`
}

type TargetingRule struct {
	CampaignID string `bson:"campaign_id" json:"campaign_id"`
	Rules      []Rule `bson:"rules" json:"rules"`
}
