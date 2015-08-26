package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry-incubator/bbs/db/codec"
)

type ETCDFlags struct {
	certFile    string
	keyFile     string
	caFile      string
	clusterUrls string
	encoding    string
}

type ETCDOptions struct {
	CertFile    string
	KeyFile     string
	CAFile      string
	ClusterUrls []string
	IsSSL       bool
	Encoding    codec.Kind
}

func AddETCDFlags(flagSet *flag.FlagSet) *ETCDFlags {
	flags := &ETCDFlags{}

	flagSet.StringVar(
		&flags.clusterUrls,
		"etcdCluster",
		"http://127.0.0.1:4001",
		"comma-separated list of etcd URLs (scheme://ip:port)",
	)

	flagSet.StringVar(
		&flags.certFile,
		"etcdCertFile",
		"",
		"Location of the client certificate for mutual auth",
	)
	flagSet.StringVar(
		&flags.keyFile,
		"etcdKeyFile",
		"",
		"Location of the client key for mutual auth",
	)
	flagSet.StringVar(
		&flags.caFile,
		"etcdCaFile",
		"",
		"Location of the CA certificate for mutual auth",
	)
	flagSet.StringVar(
		&flags.encoding,
		"etcdEncoding",
		"",
		"none,unencoded,base64",
	)
	return flags
}

func (flags *ETCDFlags) Validate() (*ETCDOptions, error) {
	scheme := ""
	clusterUrls := strings.Split(flags.clusterUrls, ",")
	for i, uString := range clusterUrls {
		uString = strings.TrimSpace(uString)
		clusterUrls[i] = uString
		u, err := url.Parse(uString)
		if err != nil {
			return nil, fmt.Errorf("Invalid cluster URL: '%s', error: [%s]", uString, err.Error())
		}
		if scheme == "" {
			if u.Scheme != "http" && u.Scheme != "https" {
				return nil, errors.New("Invalid scheme: " + uString)
			}
			scheme = u.Scheme
		} else if scheme != u.Scheme {
			return nil, fmt.Errorf("Multiple url schemes provided: %s", flags.clusterUrls)
		}
	}

	isSSL := false
	if scheme == "https" {
		isSSL = true
		if flags.certFile == "" {
			return nil, errors.New("Cert file must be provided for https connections")
		}
		if flags.keyFile == "" {
			return nil, errors.New("Key file must be provided for https connections")
		}
	}

	var encoding codec.Kind
	switch flags.encoding {
	case "", "none":
		encoding = codec.NONE
	case "unencoded":
		encoding = codec.UNENCODED
	case "base64":
		encoding = codec.BASE64
	default:
		return nil, errors.New("Encoding must be set to none, unencoded, or base64")
	}

	return &ETCDOptions{
		CertFile:    flags.certFile,
		KeyFile:     flags.keyFile,
		CAFile:      flags.caFile,
		ClusterUrls: clusterUrls,
		IsSSL:       isSSL,
		Encoding:    encoding,
	}, nil
}
