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
	if pr == nil {
		return nil
	}

	r := Routes{}
	for k, v := range pr.Routes {
		raw := json.RawMessage(v)
		r[k] = &raw
	}

	return &r
}

// func (r *Routes) Marshal() ([]byte, error) {
// 	return proto.Marshal(r.ToProto())
// }

// func (r *Routes) Unmarshal(b []byte) error {
// 	return proto.Unmarshal(b, r.ToProto())
// }

// func (r *Routes) MarshalTo(data []byte) (n int, err error) {
// 	return r.protoRoutes().MarshalTo(data)
// }

// func (r *Routes) Unmarshal(data []byte) error {
// 	pr := &ProtoRoutes{}
// 	err := pr.Unmarshal(data)
// 	if err != nil {
// 		return err
// 	}

// 	if pr.Routes == nil {
// 		return nil
// 	}

// 	routes := map[string]*json.RawMessage{}
// 	for k, v := range pr.Routes {
// 		raw := json.RawMessage(v)
// 		routes[k] = &raw
// 	}
// 	*r = routes

// 	return nil
// }

// func (r *Routes) Size() int {
// 	if r == nil {
// 		return 0
// 	}

// 	return r.protoRoutes().Size()
// }

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
