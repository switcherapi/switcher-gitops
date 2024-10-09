package controller

import (
	"net/http"
	"os"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/core"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

type ApiController struct {
	coreHandler       *core.CoreHandler
	routeCheckApiPath string
	routeApiDocsPath  string
}

type ApiCheckResponse struct {
	Status      string              `json:"status"`
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
		coreHandler:       coreHandler,
		routeCheckApiPath: "/api/check",
		routeApiDocsPath:  "/api/docs",
	}
}

func (controller *ApiController) RegisterRoutes(r *mux.Router) http.Handler {
	r.Use(DefaultHeaders)
	r.NewRoute().Path(controller.routeCheckApiPath).Name("CheckApi").HandlerFunc(controller.CheckApiHandler).Methods(http.MethodGet)
	r.NewRoute().Path(controller.routeApiDocsPath).Name("ApiDocs").HandlerFunc(controller.ApiDocsHandler).Methods(http.MethodGet)

	return r
}

func (controller *ApiController) CheckApiHandler(w http.ResponseWriter, r *http.Request) {
	utils.ResponseJSON(w, ApiCheckResponse{
		Status:      "All good",
		Version:     "1.0.1",
		ReleaseTime: config.GetEnv("RELEASE_TIME"),
		ApiSettings: ApiSettingsResponse{
			SwitcherURL:       config.GetEnv("SWITCHER_API_URL"),
			SwitcherSecret:    len(config.GetEnv("SWITCHER_API_JWT_SECRET")) > 0,
			GitTokenSecret:    len(config.GetEnv("GIT_TOKEN_PRIVATE_KEY")) > 0,
			CoreHandlerStatus: controller.coreHandler.Status,
			NumGoroutines:     runtime.NumGoroutine(),
		},
	}, http.StatusOK)
}

func (controller *ApiController) ApiDocsHandler(w http.ResponseWriter, r *http.Request) {
	fileContent, _ := os.ReadFile("./resources/swagger.yaml")

	w.WriteHeader(http.StatusOK)
	w.Write(fileContent)
}
