package models

import (
	"encoding/json"
)

type Routes map[string]json.RawMessage

func (r *Routes) protoRoutes() *ProtoRoutes {
	pr := &ProtoRoutes{
		Routes: map[string][]byte{},
	}

	for k, v := range *r {
		pr.Routes[k] = v
	}

	return pr
}

func (r *ProtoRoutes) routes() *Routes {
	nr := Routes{}

	for k, v := range r.Routes {
		var rawMessage json.RawMessage
		rawMessage = v
		nr[k] = rawMessage
	}

	return &nr
}

func (pr *ProtoRoutes) UnmarshalJSON(data []byte) error {
	tempRoutes := Routes{}
	err := json.Unmarshal(data, &tempRoutes)
	if err != nil {
		return err
	}

	byteMap := make(map[string][]byte)
	for k, v := range tempRoutes {
		byteMap[k] = []byte(v)
	}

	pr.Routes = byteMap

	return nil
}

func (r *ProtoRoutes) MarshalJSON() ([]byte, error) {
	return r.routes().Marshal()
}

func (r *Routes) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// func (r *Routes) MarshalTo(data []byte) (n int, err error) {
// 	return r.protoRoutes().MarshalTo(data)
// }

// func (r *Routes) Unmarshal(data []byte) error {
// 	return json.Unmarshal(data, r.protoRoutes().Routes)
// }

// func (r *Routes) Size() int {
// 	if r == nil {
// 		return 0
// 	}

// 	return r.protoRoutes().Size()
// }

// func (r *Routes) Equal(other Routes) bool {
// 	for k, v := range *r {
// 		if !bytes.Equal(*v, *other[k]) {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (r Routes) Validate() error {
// 	totalRoutesLength := 0
// 	if r != nil {
// 		for _, value := range r {
// 			totalRoutesLength += len(*value)
// 			if totalRoutesLength > maximumRouteLength {
// 				return ErrInvalidField{"routes"}
// 			}
// 		}
// 	}
// 	return nil
// }

func (r *ProtoRoutes) Validate() error {
	totalRoutesLength := 0
	if r != nil {
		for _, value := range r.Routes {
			totalRoutesLength += len(value)
			if totalRoutesLength > maximumRouteLength {
				return ErrInvalidField{"routes"}
			}
		}
	}
	return nil
}
