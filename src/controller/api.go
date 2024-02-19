package controller

import (
	"net/http"

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

func (controller *ApiController) CheckApiHandler(w http.ResponseWriter, r *http.Request) {
	utils.ResponseJSON(w, ApiCheckResponse{Message: "API is working"}, http.StatusOK)
}
