package stores

import (
	"context"
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/models"
)

type TargetingRule struct {
	CampaignID string `bson:"campaign_id"`
	Rules      []Rule `bson:"rules"`
}

type Rule struct {
	Dimension string   `bson:"dimension"`
	Include   []string `bson:"include"`
	Exclude   []string `bson:"exclude"`
}

type Campaign struct {
	CampaignID string `bson:"campaign_id"`
	Name       string `bson:"name"`
	Image      string `bson:"image"`
	CTA        string `bson:"cta"`
	Status     string `bson:"status"`
}

func insertRules(collection *mongo.Collection) {
	file, err := os.Open("./testdata/rules.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var allDimensions = []string{"country", "os", "app"}

	// Map to hold CampaignID -> Rules
	campaignMap := make(map[string][]Rule)

	// Parse CSV rows into rules
	for _, record := range records[1:] { // Skip header
		campaignID := record[0]
		dimension := record[1]
		include := strings.Split(record[2], "|")
		exclude := strings.Split(record[3], "|")

		// Ensure that even empty dimensions are inserted as empty arrays
		if len(include) == 1 && include[0] == "" {
			include = []string{} // Insert empty array
		}
		if len(exclude) == 1 && exclude[0] == "" {
			exclude = []string{} // Insert empty array
		}

		// Create rule
		rule := Rule{
			Dimension: dimension,
			Include:   include,
			Exclude:   exclude,
		}

		// Append the rule to the corresponding campaign
		campaignMap[campaignID] = append(campaignMap[campaignID], rule)
	}

	// Add empty dimensions if missing
	for campaignID, rules := range campaignMap {
		existingDimensions := make(map[string]bool)
		for _, rule := range rules {
			existingDimensions[rule.Dimension] = true
		}

		// For each possible dimension, if it's missing, add an empty rule
		for _, dimension := range allDimensions {
			if !existingDimensions[dimension] {
				emptyRule := Rule{
					Dimension: dimension,
					Include:   []string{},
					Exclude:   []string{},
				}
				rules = append(rules, emptyRule)
			}
		}

		// Update the campaignMap with the full rule set
		campaignMap[campaignID] = rules
	}

	// Insert each campaign's rules into MongoDB
	for campaignID, rules := range campaignMap {
		campaignRule := TargetingRule{
			CampaignID: campaignID,
			Rules:      rules,
		}

		// Insert into MongoDB, ensuring we do not insert duplicates
		filter := bson.M{"campaign_id": campaignID}
		update := bson.M{"$setOnInsert": campaignRule}
		_, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Fatalf("Failed to insert campaign %s: %v", campaignID, err)
		}
	}

}

func insertCampaigns(collection *mongo.Collection) {
	file, err := os.Open("./testdata/campaigns.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for i, record := range records {
		if i == 0 {
			continue
		}

		campaign := Campaign{
			CampaignID: record[0],
			Name:       record[1],
			Image:      record[2],
			CTA:        record[3],
			Status:     record[4],
		}

		// Upsert: Insert if not exists or update if exists
		filter := bson.M{"campaign_id": campaign.CampaignID}
		_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": campaign}, options.Update().SetUpsert(true))
		if err != nil {
			log.Printf("Failed to upsert campaign %v: %v", campaign, err)
		}
	}
}

func setupStore(t *testing.T) *Store {
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DB_NAME", "delivery_service")

	helper := helpers.New()

	store := New(helper.DB, helper.Redis, helper.Logger, helper.Metrics.CacheHits, helper.Metrics.CacheMisses)
	insertRules(store.ruleCollection)
	insertCampaigns(store.campaignCollection)

	return &store
}

func TestStore_Get(t *testing.T) {
	store := setupStore(t)

	tests := []struct {
		name              string
		dimensions        *models.Dimension
		cacheData         string
		expectedCampaigns []models.Response
		expectedErr       error
	}{
		{
			name:              "Cache Hit for Spotify",
			dimensions:        &models.Dimension{APPID: "spotify", OS: "iOS", Country: "us"},
			cacheData:         `[{"cid": "1", "name": "Spotify Campaign", "img": "image1.png", "cta": "Download", "status": "ACTIVE"}]`,
			expectedCampaigns: []models.Response{{CampaignID: "1", Image: "image1.png", CTA: "Download"}},
			expectedErr:       nil,
		},
		{
			name:       "Cache Miss - Fetch from MongoDB",
			dimensions: &models.Dimension{APPID: "exampleApp", OS: "android", Country: "us"},
			cacheData:  "",
			expectedCampaigns: []models.Response{
				{CampaignID: "spotify", Image: "https://example.com/images/spotify.png", CTA: "Listen Now"},
			}, expectedErr: nil,
		},
		{
			name:              "No Campaigns Found",
			dimensions:        &models.Dimension{APPID: "nonExistentApp", OS: "windows", Country: "northkorea"},
			cacheData:         "",
			expectedCampaigns: nil,
			expectedErr:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cacheData != "" {
				store.redisClient.Set(context.Background(), generateCacheKey(tt.dimensions.APPID, tt.dimensions.OS,
					tt.dimensions.Country), tt.cacheData, 1*time.Second)
			}

			result, err := store.Get(&gin.Context{}, tt.dimensions)

			assert.Equal(t, tt.expectedErr, err)
			if result != nil {
				assert.ElementsMatch(t, tt.expectedCampaigns, *result)
			} else {
				assert.ElementsMatch(t, tt.expectedCampaigns, result)
			}
		})
	}
}

func TestStore_InvalidateCache(t *testing.T) {
	store := setupStore(t)

	tests := []struct {
		name         string
		campaignID   string
		mockResponse error
		expectedErr  error
	}{
		{
			name:        "Successful Cache Invalidation",
			campaignID:  "campaign:amazonprime:keys",
			expectedErr: nil,
		},
		{
			name:        "Error while Invalidating Cache",
			campaignID:  "campaign_2",
			expectedErr: &helpers.Error{Code: "Internal Server Error", StatusCode: http.StatusInternalServerError, Reason: "redis: client is closed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.campaignID == "campaign_2" {
				store.redisClient.Close()
			}

			err := store.InvalidateCampaignCache(&gin.Context{}, tt.campaignID)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
