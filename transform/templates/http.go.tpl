package http

func (c *Client) Do(req *Request) (*Response, error) {
	start := time.Now();
	resp, err := c.do(req);
	duration := time.Since(start);
	fmt.Println("duration", duration);

	return resp, err
}