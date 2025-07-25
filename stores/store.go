package stores

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/models"
)

type Store struct {
	redisClient        *redis.Client
	logger             *slog.Logger
	ruleCollection     *mongo.Collection
	campaignCollection *mongo.Collection
	cacheHit           *prometheus.CounterVec
	cacheMiss          *prometheus.CounterVec
}

func New(db *mongo.Database, redisClient *redis.Client, logger *slog.Logger, cacheHit, cacheMiss *prometheus.CounterVec) Store {
	ruleCollection := db.Collection("rules")
	campaignCollection := db.Collection("campaigns")

	return Store{ruleCollection: ruleCollection, campaignCollection: campaignCollection, redisClient: redisClient,
		logger: logger, cacheHit: cacheHit, cacheMiss: cacheMiss}
}

func (s *Store) Get(ctx *gin.Context, dimensions *models.Dimension) (*[]models.Response, error) {
	cacheKey := generateCacheKey(dimensions.APPID, dimensions.OS, dimensions.Country)

	cachedCampaigns, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		s.cacheMiss.WithLabelValues("campaigns").Inc()
	} else if err == nil {
		s.cacheHit.WithLabelValues("campaigns").Inc()
		var campaigns []models.Response
		if err := json.Unmarshal([]byte(cachedCampaigns), &campaigns); err == nil {
			return &campaigns, nil
		}
	}

	ruleFilter := bson.M{
		"$and": []bson.M{
			createDimensionRule("app", dimensions.APPID),
			createDimensionRule("country", dimensions.Country),
			createDimensionRule("os", dimensions.OS),
		},
	}

	cur, err := s.ruleCollection.Find(ctx, ruleFilter)
	if err != nil {
		s.logger.Error("Error while Fetching Rules", "Error", err.Error())
		return nil, &helpers.Error{Code: "Internal Server Error", StatusCode: http.StatusInternalServerError, Reason: err.Error()}
	}
	defer cur.Close(ctx)

	var campaignIDs []string
	for cur.Next(ctx) {
		var rule models.TargetingRule
		if err := cur.Decode(&rule); err != nil {
			s.logger.Error("Error decoding rule:", "Error", err.Error())
			continue
		}
		campaignIDs = append(campaignIDs, rule.CampaignID)
	}

	if len(campaignIDs) == 0 {
		return nil, nil
	}

	freshCampaigns, err := s.FindActiveCampaignsByIDs(ctx, campaignIDs)
	if err != nil {
		return nil, err
	}

	campaignJSON, err := json.Marshal(freshCampaigns)
	if err == nil {
		s.redisClient.Set(ctx, cacheKey, campaignJSON, 10*time.Hour)

		for _, campaignID := range campaignIDs {
			err = s.redisClient.SAdd(ctx, "campaign:"+campaignID+":keys", cacheKey).Err()
			if err != nil {
				s.logger.Error("Error storing cache key in Redis set", "Error", err.Error())
			}
		}
	}

	return freshCampaigns, nil
}

func generateCacheKey(appID, os, country string) string {
	return "campaign:" + appID + ":" + os + ":" + country
}

func createDimensionRule(dimension, value string) bson.M {
	return bson.M{
		"$or": []bson.M{
			{
				"rules": bson.M{
					"$not": bson.M{
						"$elemMatch": bson.M{
							"dimension": dimension,
							"$or": []bson.M{
								{"include": bson.M{"$exists": true, "$ne": []string{}}},
								{"exclude": bson.M{"$exists": true, "$ne": []string{}}},
							},
						},
					},
				},
			},
			{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"dimension": dimension,
						"include": bson.M{
							"$exists": true,
							"$ne":     []string{},
							"$in":     []string{value},
						},
					},
				},
			},
			{
				"rules": bson.M{
					"$elemMatch": bson.M{
						"dimension": dimension,
						"exclude": bson.M{
							"$exists": true,
							"$ne":     []string{},
							"$nin":    []string{value},
						},
					},
				},
			},
		},
	}
}

func (s *Store) FindActiveCampaignsByIDs(ctx context.Context, campaignIDs []string) (*[]models.Response, error) {
	filter := bson.M{
		"campaign_id": bson.M{"$in": campaignIDs},
		"status":      "ACTIVE",
	}

	var campaigns []models.Response
	cur, err := s.campaignCollection.Find(ctx, filter)
	if err != nil {
		s.logger.Error("Error while Fetching campaigns", "Error", err.Error())
		return nil, &helpers.Error{Code: "Internal Server Error", StatusCode: http.StatusInternalServerError, Reason: err.Error()}
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var campaign models.Response
		err := cur.Decode(&campaign)
		if err != nil {
			s.logger.Error("Error decoding campaign", "Error", err.Error())
			continue
		}
		campaigns = append(campaigns, campaign)
	}

	if err := cur.Err(); err != nil {
		return nil, &helpers.Error{Code: "Internal Server Error", StatusCode: http.StatusInternalServerError, Reason: err.Error()}
	}

	return &campaigns, nil
}

func (s *Store) InvalidateCampaignCache(ctx context.Context, campaignID string) error {
	cacheKeys, err := s.redisClient.SMembers(ctx, "campaign:"+campaignID+":keys").Result()
	if err != nil {
		s.logger.Error("Error fetching cache keys from Redis set", "campaignID", campaignID, "Error", err.Error())
		return &helpers.Error{Code: "Internal Server Error", StatusCode: http.StatusInternalServerError, Reason: err.Error()}
	}

	if len(cacheKeys) == 0 {
		s.logger.Info("No cache keys found for campaign", "campaignID", campaignID)
		return nil
	}

	for _, cacheKey := range cacheKeys {
		err = s.redisClient.Del(ctx, cacheKey).Err()
		if err != nil {
			s.logger.Error("Error deleting cache key", "cacheKey", cacheKey, "Error", err.Error())
			continue
		}
		s.logger.Info("Cache key invalidated", "cacheKey", cacheKey)
	}

	err = s.redisClient.Del(ctx, "campaign:"+campaignID+":keys").Err()
	if err != nil {
		s.logger.Error("Error deleting Redis set for campaign", "campaignID", campaignID, "Error", err.Error())
		return err
	}

	s.logger.Info("Cache invalidated for campaign", "campaignID", campaignID)
	return nil
}
