# Use OpenBridger with Claude Code

Routes that support the Anthropic-compatible protocol can connect through environment variables.

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
cd /path/to/your/project
claude
```

Restore the official service with:

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN
```

For login prompts, reopen the terminal and verify both variables. For 401 errors, recopy the key. For model errors, confirm that your group can use the requested model.

<div class="callout warning"><strong>Compatibility</strong><p>The model, route, and Anthropic protocol must all match.</p></div>
