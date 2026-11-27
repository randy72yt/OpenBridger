# Usage and Logs

Open [Usage Logs]({{CONSOLE_URL}}/usage-logs) to verify whether requests reached OpenBridger, which model ran, how many tokens were used, and the exact cost.

## Key Columns

| Column | What to Check |
| --- | --- |
| **Time & Key Name** | Give each client its own API key (e.g., `codex`, `cherry`) so you can filter usage per tool. |
| **Model ID** | The actual model routed by the gateway. |
| **Prompt / Completion Tokens** | Input tokens (including conversation history) and output tokens (including reasoning tokens). |
| **Cost (¥)** | Final settled charge in RMB. |
| **Request ID** | Matches the `X-Oneapi-Request-Id` header — share this ID (never your full API key) when contacting support. |

- **Async image & video jobs**: Check the Task / Drawing log tabs rather than standard chat logs.
- **Why a failed attempt isn't listed**: Requests that 404 before hitting a valid route or fail pre-hold (`403` insufficient quota) are not billed and won't create a charge entry.

## Related

- [Credits and billing](/guide/billing/)
- [Troubleshooting](/troubleshooting/)
- [Error codes](/guide/errors/)
