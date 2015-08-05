package models

const (
	maximumAnnotationLength = 10 * 1024
	maximumRouteLength      = 4 * 1024
)

type ContainerRetainment int

const (
	_ ContainerRetainment = iota
	KeepContainer
	DeleteContainer
)
