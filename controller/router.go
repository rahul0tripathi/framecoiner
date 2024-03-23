package controller

import (
	v1 "github.com/rahul0tripathi/framecoiner/controller/v1"
	"github.com/rahul0tripathi/framecoiner/pkg/server"
)

func SetupRouter(router server.Router) {
	handler := v1.NewHandler()

	router.GET("/v1", handler.MakeGetFrameCoinerMetadataHandler())
}
