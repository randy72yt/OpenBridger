# Usage and logs

Usage Logs answer four questions: did the request reach OpenBridger, which model ran, how much was charged, and what status came back on failure.

URL: [{{CONSOLE_URL}}/usage-logs]({{CONSOLE_URL}}/usage-logs).

## Fields you can match

- Time
- Key name
- Model ID
- Charged credit
- Request id (response header `X-Oneapi-Request-Id`; the error body also ends with `(request id: …)`)

Drawing and other async tasks are in the task or drawing logs (`/usage-logs/task`, `/usage-logs/drawing`), not only in the chat log.

## Practice

- Give each client its own key, and filter by that key instead of only by model name.
- A name shown in the client can be the client's alias. The model ID in the log is the one that was billed.
- When you report a failure, send the request id, time with timezone, model ID, and status. Send at most the first 8 characters of the key, for example `sk-abcd`.
- A sudden high charge usually means the full history was sent again, or an agent called tools in a loop.

## Reserves and logs

The log row is written after settlement. A request that fails during reserve (for example insufficient credit) may have no success row. The error body still has a request id. Keep that body.

## Related

- [Credits and billing](/guide/billing/)
- [Troubleshooting](/troubleshooting/)
- [Error codes](/guide/errors/)
