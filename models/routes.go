package models

import (
	bytes "bytes"
	"encoding/json"
)

type Routes map[string]*json.RawMessage

func ParseRoutes(b map[string][]byte) *Routes {
	if b == nil {
		return nil
	}

	routes := Routes{}
	for k, v := range b {
		raw := json.RawMessage(v)
		routes[k] = &raw
	}

	return &routes
}

func (r *Routes) ToProto() *ProtoRoutes {
	if r == nil {
		return nil
	}
	pr := &ProtoRoutes{
		Routes: map[string][]byte{},
	}
	// pr := make(map[string][]byte)

	for k, v := range *r {
		pr.Routes[k] = *v
	}

	return pr
}

func (pr *ProtoRoutes) FromProto() *Routes {
	if pr == nil || pr.Routes == nil || len(pr.Routes) == 0 {
		return nil
	}

	r := Routes{}
	for k, v := range pr.Routes {
		raw := json.RawMessage(v)
		r[k] = &raw
	}

	return &r
}

func (r *Routes) Equal(other Routes) bool {
	if other == nil {
		return r == nil
	}
	for k, v := range *r {
		if !bytes.Equal(*v, *other[k]) {
			return false
		}
	}
	return true
}

func (r Routes) Validate() error {
	totalRoutesLength := 0
	for _, value := range r {
		totalRoutesLength += len(*value)
		if totalRoutesLength > maximumRouteLength {
			return ErrInvalidField{"routes"}
		}
	}
	return nil
}
