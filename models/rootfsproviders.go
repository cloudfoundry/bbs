package models

type RootFSProviders map[string]*Providers

func (r *RootFSProviders) protoRootFSProviders() *ProtoRootfsproviders {
	pr := &ProtoRootfsproviders{
		RootfsProviders: map[string]*Providers{},
	}

	for k, v := range *r {
		pr.RootfsProviders[k] = v
	}

	return pr
}

func (r *RootFSProviders) Marshal() ([]byte, error) {
	return r.protoRootFSProviders().Marshal()
}

func (r *RootFSProviders) MarshalTo(data []byte) (n int, err error) {
	return r.protoRootFSProviders().MarshalTo(data)
}

func (r *RootFSProviders) Unmarshal(data []byte) error {
	pr := &ProtoRootfsproviders{}
	err := pr.Unmarshal(data)
	if err != nil {
		return err
	}

	if pr.RootfsProviders == nil {
		return nil
	}

	rootFSProviders := map[string]*Providers{}
	for k, v := range pr.RootfsProviders {
		rootFSProviders[k] = v
	}
	*r = rootFSProviders

	return nil
}

func (r *RootFSProviders) Size() int {
	if r == nil {
		return 0
	}

	return r.protoRootFSProviders().Size()
}

func (r *RootFSProviders) Equal(other RootFSProviders) bool {
	for k, v := range *r {
		oSlice := *other[k]
		for i, str1 := range v.GetProvidersList() {
			if str1 != oSlice.GetProvidersList()[i] {
				return false
			}
		}
	}
	return true
}

func (r RootFSProviders) Validate() error {
	totalRootFSProvidersLength := 0
	if r != nil {
		for _, value := range r {
			totalRootFSProvidersLength += len(value.GetProvidersList())
			if totalRootFSProvidersLength > maximumRouteLength {
				return ErrInvalidField{"rootFSProviders"}
			}
		}
	}
	return nil
}
