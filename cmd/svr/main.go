package main

import (
	"content-service-api/pkg/api"

	"github.com/sirupsen/logrus"
)

func main() {
	if err := api.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("Error serving API")
	}
}
