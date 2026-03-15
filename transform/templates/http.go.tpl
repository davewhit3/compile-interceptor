package http

func (c *Client) Do(req *Request) (*Response, error) {
	start := time.Now();
	resp, err := c.do(req);
	duration := time.Since(start);

	fmt.Printf("%s %d %v %s\n", req.Method, resp.StatusCode, duration, req.URL.String())

	return resp, err
}