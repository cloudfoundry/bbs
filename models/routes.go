package models

import (
	"bytes"
	"encoding/json"
)

func (r *Routes) routes() *Routes {
	pr := &Routes{
		Routes: map[string][]byte{},
	}

	for k, v := range *r {
		pr.Routes[k] = *v
	}

	return pr
}

func (r *Routes) Marshal() ([]byte, error) {
	return r.routes().Marshal()
}

func (r *Routes) MarshalTo(data []byte) (n int, err error) {
	return r.routes().MarshalTo(data)
}

func (r *Routes) Unmarshal(data []byte) error {
	pr := &Routes{}
	err := pr.Unmarshal(data)
	if err != nil {
		return err
	}

	if pr.Routes == nil {
		return nil
	}

	routes := map[string][]byte{}
	for k, v := range pr.Routes {
		raw := json.RawMessage(v)
		routes[k] = raw
	}
	*r = routes

	return nil
}

func (r *Routes) Size() int {
	if r == nil {
		return 0
	}

	return r.routes().Size()
}

func (r *Routes) Equal(other Routes) bool {
	for k, v := range *r {
		if !bytes.Equal(*v, *other[k]) {
			return false
		}
	}
	return true
}

func (r Routes) Validate() error {
	totalRoutesLength := 0
	if r != nil {
		for _, value := range r {
			totalRoutesLength += len(*value)
			if totalRoutesLength > maximumRouteLength {
				return ErrInvalidField{"routes"}
			}
		}
	}
	return nil
}
