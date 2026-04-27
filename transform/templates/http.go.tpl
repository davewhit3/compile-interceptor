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

    var bodyStr string
    contentType := req.Header.Get("Content-Type")
    printable := []string{"application/json", "text/plain", "application/x-www-form-urlencoded"}
	for _, t := range printable {
		if strings.Contains(contentType, t) {
			bodyStr = string(bodyBytes)
			break
		}
	}

    outgoing.AddRequest(req.Method, req.URL.String(), code, duration, bodyStr)

	return resp, err
}