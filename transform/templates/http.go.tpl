package http

func (c *Client) Do(req *Request) (*Response, error) {
	start := time.Now();
	var code = -1;

	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	resp, err := c.do(req);
	duration := time.Since(start);

	if err == nil {
		code = resp.StatusCode
	}

    var sb strings.Builder

    sb.WriteString(fmt.Sprintf("method=%s code=%d dur=%v url=%s ", req.Method, code, duration, req.URL.String()))

    contentType := req.Header.Get("Content-Type")
    printable := []string{"application/json", "text/plain", "application/x-www-form-urlencoded"}
	isPrintable := false
	for _, t := range printable {
		if strings.Contains(contentType, t) {
			isPrintable = true
			break
		}
	}

    if isPrintable {
        sb.WriteString(fmt.Sprintf("body=%s", string(bodyBytes)))
    }

    sb.WriteString("\n")

    outgoing.Add(sb.String())

	return resp, err
}