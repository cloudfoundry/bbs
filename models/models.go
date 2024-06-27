package models

//go:generate bash ../scripts/generate_go_protos.sh

const (
	maximumAnnotationLength = 10 * 1024
	maximumRouteLength      = 128 * 1024
)
