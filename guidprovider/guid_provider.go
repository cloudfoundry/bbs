package guidprovier

//go:generate counterfeiter . GUIDProvider

type GUIDProvider interface {
	NextGUID() (string, error)
}
