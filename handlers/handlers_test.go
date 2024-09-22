package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Durga-Chikkala/delivery-service/constants"
	"github.com/Durga-Chikkala/delivery-service/helpers"
	"github.com/Durga-Chikkala/delivery-service/models"
	"github.com/Durga-Chikkala/delivery-service/services"
)

func TestHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDelivery := services.NewMockDelivery(ctrl)

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockCalls      []interface{}
		expectedStatus int
	}{
		{
			name: "successful response",
			queryParams: map[string]string{
				constants.App:     "com.app.test",
				constants.Country: "US",
				constants.Os:      "Android",
			},
			mockCalls: []interface{}{
				mockDelivery.EXPECT().Get(gomock.Any(), &models.Dimension{APPID: "com.app.test", Country: "US", OS: "Android"}).
					Return(&[]models.Response{{CampaignID: "spotify"}, {CampaignID: "zepto"}}, nil),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service returns no campaigns",
			queryParams: map[string]string{
				constants.App:     "com.app.test",
				constants.Country: "US",
				constants.Os:      "Android",
			},
			mockCalls: []interface{}{
				mockDelivery.EXPECT().Get(gomock.Any(), &models.Dimension{APPID: "com.app.test", Country: "US", OS: "Android"}).
					Return(nil, nil),
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "missing app parameter",
			queryParams: map[string]string{
				constants.Country: "US",
				constants.Os:      "Android",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing country parameter",
			queryParams: map[string]string{
				constants.App: "com.app.test",
				constants.Os:  "Android",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing os parameter",
			queryParams: map[string]string{
				constants.App:     "com.app.test",
				constants.Country: "US",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service returns error",
			queryParams: map[string]string{
				constants.App:     "com.app.test",
				constants.Country: "US",
				constants.Os:      "Android",
			},
			mockCalls: []interface{}{
				mockDelivery.EXPECT().Get(gomock.Any(), &models.Dimension{APPID: "com.app.test", Country: "US", OS: "Android"}).
					Return(nil, &helpers.Error{StatusCode: http.StatusInternalServerError}),
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(mockDelivery)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			query := url.Values{}
			for key, value := range tt.queryParams {
				query.Set(key, value)
			}

			c.Request = httptest.NewRequest("GET", "/dummy?"+query.Encode(), nil)

			handler.Get(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
