package main

import (
	"fmt"

	garage "git.deuxfleurs.fr/garage-sdk/garage-admin-sdk-golang"
)

type GarageClient struct {
	Client *garage.APIClient
}

func NewGarageClient(scheme, host, token string) (*GarageClient, error) {
	cfg := garage.NewConfiguration()
	cfg.Scheme = scheme
	cfg.Host = host
	cfg.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", token)

	client := garage.NewAPIClient(cfg)
	return &GarageClient{Client: client}, nil
}

