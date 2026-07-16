package httpapi

import "net/http"

func NewHandler() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/health", healthHandler)

	return router
}

func healthHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	_, _ = response.Write([]byte(`{"status":"ok"}`))
}
