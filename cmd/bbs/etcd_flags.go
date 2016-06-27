package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/bbs/db/etcd"
)

type ETCDFlags struct {
	etcdCertFile           string
	etcdKeyFile            string
	etcdCaFile             string
	clusterUrls            string
	clientSessionCacheSize int
	maxIdleConnsPerHost    int
}

func AddETCDFlags(flagSet *flag.FlagSet) *ETCDFlags {
	flags := &ETCDFlags{}

	flagSet.StringVar(
		&flags.clusterUrls,
		"etcdCluster",
		"",
		"comma-separated list of etcd URLs (scheme://ip:port)",
	)
	flagSet.StringVar(
		&flags.etcdCertFile,
		"etcdCertFile",
		"",
		"Location of the client certificate for mutual auth",
	)
	flagSet.StringVar(
		&flags.etcdKeyFile,
		"etcdKeyFile",
		"",
		"Location of the client key for mutual auth",
	)
	flagSet.StringVar(
		&flags.etcdCaFile,
		"etcdCaFile",
		"",
		"Location of the CA certificate for mutual auth",
	)

	flagSet.IntVar(
		&flags.clientSessionCacheSize,
		"etcdClientSessionCacheSize",
		0,
		"Capacity of the ClientSessionCache option on the TLS configuration. If zero, golang's default will be used",
	)
	flagSet.IntVar(
		&flags.maxIdleConnsPerHost,
		"etcdMaxIdleConnsPerHost",
		0,
		"Controls the maximum number of idle (keep-alive) connctions per host. If zero, golang's default will be used",
	)
	return flags
}

func (flags *ETCDFlags) Validate() (*etcd.ETCDOptions, error) {
	if flags.clusterUrls == "" && flags.etcdCaFile == "" &&
		flags.etcdKeyFile == "" && flags.etcdCertFile == "" {
		return &etcd.ETCDOptions{IsConfigured: false}, nil
	}

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
		if flags.etcdCertFile == "" {
			return nil, errors.New("Cert file must be provided for https connections")
		}
		if flags.etcdKeyFile == "" {
			return nil, errors.New("Key file must be provided for https connections")
		}
	}

	return &etcd.ETCDOptions{
		CertFile:    flags.etcdCertFile,
		KeyFile:     flags.etcdKeyFile,
		CAFile:      flags.etcdCaFile,
		ClusterUrls: clusterUrls,
		IsSSL:       isSSL,
		ClientSessionCacheSize: flags.clientSessionCacheSize,
		MaxIdleConnsPerHost:    flags.maxIdleConnsPerHost,
		IsConfigured:           true,
	}, nil
}
