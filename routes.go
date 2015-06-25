package bbs

import "github.com/tedsuo/rata"

const (
	// Domains
	DomainsRoute = "Domains"
)

var Routes = rata.Routes{
	// Domains
	{Path: "/v1/domains", Method: "GET", Name: DomainsRoute},
}
