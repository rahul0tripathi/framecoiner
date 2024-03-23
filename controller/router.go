package controller

import (
	v1 "github.com/rahul0tripathi/framecoiner/controller/v1"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
)

func SetupRouter(accountSvc v1.AccountService, router server.Router) {
	handler := v1.NewHandler()

	router.GET("/v1", handler.MakeGetFrameCoinerMetadataHandler())
	router.GET("/v1/account/:owner", handler.MakeGetAccountHandler(accountSvc))
}
