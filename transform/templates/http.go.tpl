package http

func (c *Client) Do(req *Request) (*Response, error) {
	start := time.Now();
	var code = -1;
	resp, err := c.do(req);
	duration := time.Since(start);

	if err == nil {
		code = resp.StatusCode
	}

	outgoing.Add(fmt.Sprintf("%s %d %v %s\n", req.Method, code, duration, req.URL.String()))

	return resp, err
}