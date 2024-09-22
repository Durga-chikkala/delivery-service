package helpers

func FormResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
	}
}
