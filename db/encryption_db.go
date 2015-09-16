package db

import "github.com/pivotal-golang/lager"

//go:generate counterfeiter . EncryptionDB

type EncryptionDB interface {
	EncryptionKeyLabel(logger lager.Logger) (string, error)
	SetEncryptionKeyLabel(logger lager.Logger, encryptionKeyLabel string) error
}
