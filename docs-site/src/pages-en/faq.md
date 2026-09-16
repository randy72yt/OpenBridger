# Frequently asked questions

## What Base URL should I use?

Most OpenAI-compatible clients use `{{API_BASE_URL}}`. Add `/chat/completions` only when a setting explicitly requires the complete endpoint.

## Why is the model list empty?

Some clients do not read compatible model lists or filter unknown IDs. Copy the exact model ID from the console and add it manually.

## How can I confirm that a request used OpenBridger?

Create a dedicated key, send a short message, then match the time, model, and key in Usage Logs.

## Which clients are supported?

Tools that accept a custom Base URL, API key, and model ID can usually connect. Actual capabilities depend on the protocol, model, and route.

## Which balance is charged first?

The console's account and plan rules define charge priority. Use Usage Logs to inspect the group, model, and charge source for a specific request.

## Which payment methods are supported?

Use only methods displayed in Wallet and checkout. Availability can vary by region, currency, device, and browser.

## How do I restore a tool's official models?

Switch to the built-in provider, disable the custom Base URL, clear the related environment variables, and restart the client.
