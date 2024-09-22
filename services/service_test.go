package services

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/Durga-Chikkala/delivery-service/stores"
)

func TestService_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := stores.NewMockDelivery(ctrl)
	ctx := &gin.Context{}

	service := New(mockStore)

	tests := []struct {
		name           string
		dimensions     *models.Dimension
		mockCalls      []interface{}
		expectedResult *[]models.Response
		expectedError  error
	}{
		{
			name: "successful response",
			dimensions: &models.Dimension{
				APPID:   "com.app.test",
				Country: "US",
				OS:      "Android",
			},
			mockCalls: []interface{}{
				mockStore.EXPECT().Get(ctx, &models.Dimension{
					APPID:   "com.app.test",
					Country: "us",
					OS:      "android",
				}).Return(&[]models.Response{{CampaignID: "Campaign 1"}}, nil),
			},
			expectedResult: &[]models.Response{{CampaignID: "Campaign 1"}},
			expectedError:  nil,
		},
		{
			name: "service returns error",
			dimensions: &models.Dimension{
				APPID:   "com.app.test",
				Country: "us",
				OS:      "android",
			},
			mockCalls: []interface{}{
				mockStore.EXPECT().Get(ctx, &models.Dimension{
					APPID:   "com.app.test",
					Country: "us",
					OS:      "android",
				}).Return(&[]models.Response{}, &helpers.Error{StatusCode: http.StatusInternalServerError}),
			},
			expectedResult: &[]models.Response{},
			expectedError:  &helpers.Error{StatusCode: http.StatusInternalServerError},
		},
		{
			name: "nil response from store",
			dimensions: &models.Dimension{
				APPID:   "com.app.test",
				Country: "US",
				OS:      "Android",
			},
			mockCalls: []interface{}{
				mockStore.EXPECT().Get(ctx, &models.Dimension{
					APPID:   "com.app.test",
					Country: "us",
					OS:      "android",
				}).Return(nil, nil),
			},
			expectedResult: nil,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := service.Get(ctx, tt.dimensions)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
