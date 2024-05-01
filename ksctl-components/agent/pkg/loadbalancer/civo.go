package loadbalancer

type civoLB struct {
}

func (*civoLB) Create() (*string, error) {
	return func() *string {
		v := ""
		return &v
	}(), nil
}

func (*civoLB) Delete() error {
	return nil
}
