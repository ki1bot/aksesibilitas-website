package httpapi

import "net/http"

func (server *Server) health(
	writer http.ResponseWriter,
	request *http.Request,
) {
	response := healthResponse{
		Status:   "ok",
		Database: "ok",
		Redis:    "ok",
	}

	statusCode := http.StatusOK

	if err := server.pool.Ping(
		request.Context(),
	); err != nil {
		response.Status = "degraded"
		response.Database = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	if err := server.redisClient.Ping(
		request.Context(),
	).Err(); err != nil {
		response.Status = "degraded"
		response.Redis = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(writer, statusCode, response)
}
