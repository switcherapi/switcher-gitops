package controller

import (
	"net/http"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type ApiController struct {
	CoreHandler       *core.CoreHandler
	RouteCheckApiPath string
}

type ApiCheckResponse struct {
	Status      string              `json:"message"`
	Version     string              `json:"version"`
	ReleaseTime string              `json:"release_time"`
	ApiSettings ApiSettingsResponse `json:"api_settings"`
}

type ApiSettingsResponse struct {
	SwitcherURL       string `json:"switcher_url"`
	SwitcherSecret    bool   `json:"switcher_secret"`
	GitTokenSecret    bool   `json:"git_token_secret"`
	CoreHandlerStatus int    `json:"core_handler_status"`
	NumGoroutines     int    `json:"num_goroutines"`
}

func NewApiController(coreHandler *core.CoreHandler) *ApiController {
	return &ApiController{
		CoreHandler:       coreHandler,
		RouteCheckApiPath: "/api/check",
	}
}

func (controller *ApiController) RegisterRoutes(r *mux.Router) http.Handler {
	r.NewRoute().Path(controller.RouteCheckApiPath).Name("CheckApi").HandlerFunc(controller.CheckApiHandler).Methods(http.MethodGet)

	return r
}

func (controller *ApiController) CheckApiHandler(w http.ResponseWriter, r *http.Request) {
	utils.ResponseJSON(w, ApiCheckResponse{
		Status:      "All good",
		Version:     "1.0.0",
		ReleaseTime: config.GetEnv("RELEASE_TIME"),
		ApiSettings: ApiSettingsResponse{
			SwitcherURL:       config.GetEnv("SWITCHER_API_URL"),
			SwitcherSecret:    len(config.GetEnv("SWITCHER_API_JWT_SECRET")) > 0,
			GitTokenSecret:    len(config.GetEnv("GIT_TOKEN_PRIVATE_KEY")) > 0,
			CoreHandlerStatus: controller.CoreHandler.Status,
			NumGoroutines:     runtime.NumGoroutine(),
		},
	}, http.StatusOK)
}
