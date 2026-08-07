package httpapi

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

func (server *Server) listProjects(
	writer http.ResponseWriter,
	request *http.Request,
) {
	principal := principalFromContext(
		request.Context(),
	)

	projects, err :=
		server.queries.ListProjectsByUser(
			request.Context(),
			principal.User.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, projects)
}

func (server *Server) createProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input projectRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(
		input.Description,
	)

	if len(input.Name) < 2 ||
		len(input.Name) > 120 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_name",
			"Nama project harus terdiri dari 2 sampai 120 karakter",
		)
		return
	}

	if len(input.Description) > 1000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_description",
			"Deskripsi maksimal 1000 karakter",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	project, err := queries.CreateProject(
		request.Context(),
		db.CreateProjectParams{
			ID:          uuid.New(),
			OwnerID:     principal.User.ID,
			Name:        input.Name,
			Description: input.Description,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.AddProjectMember(
		request.Context(),
		db.AddProjectMemberParams{
			ProjectID: project.ID,
			UserID:    principal.User.ID,
			Role:      "owner",
		},
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		project,
	)
}

func (server *Server) getProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	project, err := server.queries.GetProjectForUser(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, project)
}

func (server *Server) updateProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	var input projectRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(
		input.Description,
	)

	if len(input.Name) < 2 ||
		len(input.Name) > 120 ||
		len(input.Description) > 1000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_project",
			"Data project tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	project, err := server.queries.UpdateProject(
		request.Context(),
		db.UpdateProjectParams{
			ProjectID:   projectID,
			UserID:      principal.User.ID,
			Name:        input.Name,
			Description: input.Description,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, project)
}

func (server *Server) deleteProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	err := server.queries.DeleteProject(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
