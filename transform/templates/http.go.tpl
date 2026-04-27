package http

func (c *Client) Do(req *Request) (*Response, error) {
	start := time.Now();
	var code = -1;

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	resp, err := c.do(req);
	duration := time.Since(start);

	if err == nil {
		code = resp.StatusCode
	}

	printable := []string{"application/json", "text/plain", "application/x-www-form-urlencoded"}

	var bodyStr string
	reqContentType := req.Header.Get("Content-Type")
	for _, t := range printable {
		if strings.Contains(reqContentType, t) {
			bodyStr = string(bodyBytes)
			break
		}
	}

	var respBodyStr string
	if err == nil && resp.Body != nil {
		respContentType := resp.Header.Get("Content-Type")
		for _, t := range printable {
			if strings.Contains(respContentType, t) {
				var respBodyBytes []byte
				respBodyBytes, _ = io.ReadAll(resp.Body)
				resp.Body = io.NopCloser(bytes.NewBuffer(respBodyBytes))
				respBodyStr = string(respBodyBytes)
				break
			}
		}
	}

	outgoing.AddRequest(req.Method, req.URL.String(), code, duration, bodyStr, respBodyStr)

	return resp, err
}
