package stores

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mt := helpers.InitializeMongo(logger)

	store := New(mt, redisClient, logger)
	insertRules(store.ruleCollection)
	insertCampaigns(store.campaignCollection)

	return &store
}

func TestStore_Get(t *testing.T) {
	store := setupStore(t)

	tests := []struct {
		name              string
		dimensions        *models.Dimension
		cacheData         map[string]interface{}
		expectedCampaigns *[]models.Response
		expectedErr       error
	}{
		{
			name:       "Cache Hit for Spotify",
			dimensions: &models.Dimension{APPID: "spotify", Country: "us", OS: "android"},
			cacheData: map[string]interface{}{
				"campaign:spotify": models.Response{CampaignID: "spotify", Image: "https://img"},
			},
			expectedCampaigns: &[]models.Response{{CampaignID: "spotify", Image: "https://img"}},
			expectedErr:       nil,
		},
		{
			name:       "Cache Miss with DB Hit for Duolingo",
			dimensions: &models.Dimension{APPID: "duolingo", Country: "IN", OS: "android"},
			cacheData:  map[string]interface{}{},

			expectedCampaigns: &[]models.Response{{CampaignID: "duolingo",
				Image: "https://example.com/images/duolingo.png", CTA: "Start Learning"}},
			expectedErr: nil,
		},
		{
			name:              "No Campaigns Found",
			dimensions:        &models.Dimension{APPID: "netflix", Country: "england", OS: "windows"},
			cacheData:         map[string]interface{}{},
			expectedCampaigns: nil,
			expectedErr:       nil,
		},
		{
			name:       "Cache Miss for WhatsApp",
			dimensions: &models.Dimension{APPID: "com.whatsapp", Country: "brazil", OS: "android"},
			cacheData:  map[string]interface{}{},
			expectedCampaigns: &[]models.Response{{CampaignID: "whatsapp", Image: "https://example.com/images/whatsapp.png", CTA: "Send Message"},
				{CampaignID: "duolingo", Image: "https://example.com/images/duolingo.png", CTA: "Start Learning"}},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.cacheData {
				data, _ := json.Marshal(v)
				store.redisClient.Set(&gin.Context{}, k, data, 5*time.Second)
			}

			campaigns, err := store.Get(&gin.Context{}, tt.dimensions)

			assert.Equal(t, tt.expectedErr, err)
			assert.EqualValues(t, tt.expectedCampaigns, campaigns)
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
			campaignID:  "amazonprime",
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

			err := store.InvalidateCache(&gin.Context{}, tt.campaignID)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
