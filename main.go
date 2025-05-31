package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var searxngClient *SearXNGClient

func main() {
	var transport string
	var host string
	var port string
	var searxngURL string

	flag.StringVar(&transport, "t", "sse", "Transport type (stdio or sse)")
	flag.StringVar(&host, "h", "0.0.0.0", "Host of sse server")
	flag.StringVar(&port, "p", "8892", "Port of sse server")
	flag.StringVar(&searxngURL, "searxng", "http://127.0.0.1:8080", "SearXNG instance URL")
	flag.Parse()

	searxngClient = NewSearXNGClient(searxngURL)

	mcpServer := server.NewMCPServer(
		"go_mcp_server_searxng",
		"1.0.0",
	)

	searchTool := mcp.NewTool("searxng_search",
		mcp.WithDescription("Search information through SearXNG. Supports various categories and search engines."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithString("categories",
			mcp.Description("Search categories (general, images, videos, news, music, files, science, it). Multiple values separated by comma"),
		),
		mcp.WithString("engines",
			mcp.Description("Search engines (google, bing, duckduckgo, yandex, etc.). Multiple values separated by comma"),
		),
		mcp.WithString("language",
			mcp.Description("Search language (ru, en, de, fr, etc.)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number of results (default 1)"),
		),
		mcp.WithString("time_range",
			mcp.Description("Time range (day, week, month, year)"),
		),
		mcp.WithNumber("safe_search",
			mcp.Description("Safe search (0 - disabled, 1 - moderate, 2 - strict)"),
		),
	)

	mcpServer.AddTool(searchTool, searxngSearchHandler)

	enginesInfoTool := mcp.NewTool("searxng_engines_info",
		mcp.WithDescription("Get information about available SearXNG search engines and categories"),
	)

	mcpServer.AddTool(enginesInfoTool, searxngEnginesInfoHandler)

	imageSearchTool := mcp.NewTool("searxng_image_search",
		mcp.WithDescription("Specialized image search through SearXNG"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for images"),
		),
		mcp.WithString("engines",
			mcp.Description("Image search engines (google images, bing images, flickr, etc.)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number of results"),
		),
	)

	mcpServer.AddTool(imageSearchTool, searxngImageSearchHandler)

	newsSearchTool := mcp.NewTool("searxng_news_search",
		mcp.WithDescription("Specialized news search through SearXNG"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query for news"),
		),
		mcp.WithString("time_range",
			mcp.Description("Time range for news (day, week, month, year)"),
		),
		mcp.WithString("language",
			mcp.Description("News language"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number of results"),
		),
	)

	mcpServer.AddTool(newsSearchTool, searxngNewsSearchHandler)

	if transport == "sse" {
		sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(fmt.Sprintf("http://localhost:%s", port)))
		log.Printf("SSE server listening on %s:%s URL: http://127.0.0.1:%s/sse", host, port, port)
		log.Printf("Using SearXNG instance: %s", searxngURL)
		if err := sseServer.Start(fmt.Sprintf("%s:%s", host, port)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		log.Printf("Stdio server started. Using SearXNG instance: %s", searxngURL)
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

func searxngSearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	params := SearchParams{
		Query:      query,
		Categories: []string{"general"},
		Engines:    []string{"google"},
		Language:   "en",
	}

	if categories, ok := request.Params.Arguments["categories"].(string); ok && categories != "" {
		params.Categories = strings.Split(categories, ",")
		for i := range params.Categories {
			params.Categories[i] = strings.TrimSpace(params.Categories[i])
		}
	}

	if engines, ok := request.Params.Arguments["engines"].(string); ok && engines != "" {
		params.Engines = strings.Split(engines, ",")
		for i := range params.Engines {
			params.Engines[i] = strings.TrimSpace(params.Engines[i])
		}
	}

	if language, ok := request.Params.Arguments["language"].(string); ok && language != "" {
		params.Language = language
	}

	if pageFloat, ok := request.Params.Arguments["page"].(float64); ok {
		params.PageNo = int(pageFloat)
	}

	if timeRange, ok := request.Params.Arguments["time_range"].(string); ok {
		params.TimeRange = timeRange
	}

	if safeSearchFloat, ok := request.Params.Arguments["safe_search"].(float64); ok {
		params.SafeSearch = int(safeSearchFloat)
	}

	result, err := searxngClient.Search(params)
	if err != nil {
		return nil, fmt.Errorf("search error: %w", err)
	}

	response := map[string]interface{}{
		"query":             result.Query,
		"number_of_results": result.NumberOfResults,
		"results":           result.Results,
	}

	if len(result.Answers) > 0 {
		response["answers"] = result.Answers
	}
	if len(result.Suggestions) > 0 {
		response["suggestions"] = result.Suggestions
	}
	if len(result.Corrections) > 0 {
		response["corrections"] = result.Corrections
	}

	jsonResult, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("result serialization error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func searxngEnginesInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config, err := searxngClient.GetEngines()
	if err != nil {
		return nil, fmt.Errorf("error getting engines information: %w", err)
	}

	jsonResult, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("result serialization error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func searxngImageSearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	params := SearchParams{
		Query:      query,
		Categories: []string{"images"},
		Engines:    []string{"google images"},
		Language:   "en",
	}

	if engines, ok := request.Params.Arguments["engines"].(string); ok && engines != "" {
		params.Engines = strings.Split(engines, ",")
		for i := range params.Engines {
			params.Engines[i] = strings.TrimSpace(params.Engines[i])
		}
	}

	if pageFloat, ok := request.Params.Arguments["page"].(float64); ok {
		params.PageNo = int(pageFloat)
	}

	result, err := searxngClient.Search(params)
	if err != nil {
		return nil, fmt.Errorf("image search error: %w", err)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("result serialization error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func searxngNewsSearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	params := SearchParams{
		Query:      query,
		Categories: []string{"news"},
		Engines:    []string{"google news"},
		Language:   "en",
	}

	if timeRange, ok := request.Params.Arguments["time_range"].(string); ok {
		params.TimeRange = timeRange
	}

	if language, ok := request.Params.Arguments["language"].(string); ok && language != "" {
		params.Language = language
	}

	if pageFloat, ok := request.Params.Arguments["page"].(float64); ok {
		params.PageNo = int(pageFloat)
	}

	result, err := searxngClient.Search(params)
	if err != nil {
		return nil, fmt.Errorf("news search error: %w", err)
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("result serialization error: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}
