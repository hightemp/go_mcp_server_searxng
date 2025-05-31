# SearXNG MCP Server

A Model Context Protocol (MCP) server for integrating with SearXNG metasearch engine.

## Features

- **General Search**: Search across multiple categories and engines
- **Image Search**: Specialized image search functionality
- **News Search**: Time-filtered news search
- **Engine Info**: Get available search engines and categories

## Installation

```bash
go mod tidy
go build
```

## Available Tools

### searxng_search
Main search tool with support for categories, engines, language, pagination, time range, and safe search.

### searxng_image_search
Specialized image search with engine filtering and pagination.

### searxng_news_search
News search with time range filtering and language support.

### searxng_engines_info
Get information about available search engines and configuration.

## Parameters

- `-t`: Transport type (stdio/sse), default: stdio
- `-h`: Host for SSE server, default: 0.0.0.0
- `-p`: Port for SSE server, default: 8892
- `-searxng`: SearXNG instance URL, default: http://127.0.0.1:8080

## Example

```bash
# Start server with custom SearXNG instance
./go_mcp_server_searxng -searxng http://127.0.0.1:8080 -t sse -p 8892
# or cli
./go_mcp_server_searxng -searxng http://127.0.0.1:8080 -t cli
```