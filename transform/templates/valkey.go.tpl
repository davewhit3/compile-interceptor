package valkey

func (c *singleClient) Do(ctx context.Context, cmd Completed) (resp ValkeyResult) {
	attempts := 1
retry:

	start := time.Now();
	resp = c.conn.Do(ctx, cmd)
	duration := time.Since(start);

	cm := cmd.Commands()
	var cmdName, cmdKey string
	if len(cm) > 0 {
		cmdName = cm[0]
	}
	if len(cm) > 1 {
		cmdKey = cm[1]
	}
	var errStr string
	if respErr := resp.Error(); respErr != nil {
		errStr = respErr.Error()
	}
	outgoing.AddCommand(cmdName, cmdKey, duration, errStr)

	if err := resp.Error(); err != nil {
		if err == errConnExpired {
			goto retry
		}
		if c.retry && cmd.IsRetryable() && c.isRetryable(err, ctx) {
			if c.retryHandler.WaitOrSkipRetry(ctx, attempts, cmd, err) {
				attempts++
				goto retry
			}
		}
	}
	if resp.NonValkeyError() == nil { // not recycle cmds if error, since cmds may be used later in the pipe.
		cmds.PutCompleted(cmd)
	}
	return resp
}
