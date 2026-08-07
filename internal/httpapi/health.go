package httpapi

import "net/http"

func (server *Server) health(
	writer http.ResponseWriter,
	request *http.Request,
) {
	response := healthResponse{
		Status:   "ok",
		Database: "ok",
	}

	statusCode := http.StatusOK

	if err := server.pool.Ping(
		request.Context(),
	); err != nil {
		response.Status = "degraded"
		response.Database = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(
		writer,
		statusCode,
		response,
	)
}
