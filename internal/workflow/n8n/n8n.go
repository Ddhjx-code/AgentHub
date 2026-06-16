package n8n

type Client struct {
	DefaultTimeout int
}

func NewClient(defaultTimeout int) *Client {
	return &Client{DefaultTimeout: defaultTimeout}
}
