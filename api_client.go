package flareio

type ApiClient struct {
	tenantId int
	apiKey   string
}

type ApiClientOption func(*ApiClient)

func WithTenantId(tenantId int) ApiClientOption {
	return func(client *ApiClient) {
		client.tenantId = tenantId
	}
}

func NewClient(
	apiKey string,
	optionFns ...ApiClientOption,
) *ApiClient {
	c := &ApiClient{
		apiKey: apiKey,
	}
	for _, optionFn := range optionFns {
		optionFn(c)
	}
	return c
}
