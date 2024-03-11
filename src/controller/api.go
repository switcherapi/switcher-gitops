package controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type ApiController struct {
	RouteCheckApiPath string
}

type ApiCheckResponse struct {
	Message string `json:"message"`
}

func NewApiController() *ApiController {
	return &ApiController{
		RouteCheckApiPath: "/api/check",
	}
}

func (controller *ApiController) RegisterRoutes(r *mux.Router) http.Handler {
	r.HandleFunc(controller.RouteCheckApiPath, controller.CheckApiHandler).Methods(http.MethodGet)

	return r
}

func (controller *ApiController) CheckApiHandler(w http.ResponseWriter, r *http.Request) {
	utils.ResponseJSON(w, ApiCheckResponse{Message: "API is working"}, http.StatusOK)
}
