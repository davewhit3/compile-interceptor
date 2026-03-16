package valkey

func (c *singleClient) Do(ctx context.Context, cmd Completed) (resp ValkeyResult) {
	attempts := 1
retry:

	start := time.Now();
	resp = c.conn.Do(ctx, cmd)
	duration := time.Since(start);

	cm := cmd.Commands()
	outgoing.Add(fmt.Sprintf("%s %s %s\n", cm[0], duration, cm[1]))
	
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