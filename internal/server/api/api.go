package api

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type API struct {
	client.Client
}
