package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
)

const (
	// defaultModel is the GA Gemini 2.0 Flash. The previous *-exp slug is being
	// retired; override with WithModel if a newer model becomes preferred.
	defaultModel   = "gemini-2.0-flash"
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

	// analyzePrompt is the instruction sent to Gemini for game analysis.
	// It demands strict JSON output matching ports.Analysis; no prose outside the object.
	analyzePrompt = `You are a chess coach. Analyse the following PGN game and return ONLY a JSON object — no text before or after the JSON, no markdown fences.

The JSON must exactly match this schema:
{
  "summary": "<short paragraph about the game overall>",
  "mistakes": [
    {
      "ply": <integer, 1-based half-move number>,
      "move": "<move that was played, in UCI notation e.g. e2e4>",
      "severity": "<one of: inaccuracy | mistake | blunder>",
      "better": "<best alternative move in UCI notation, omit if none>",
      "comment": "<one sentence explanation>"
    }
  ]
}

Rules:
- Include at most 6 mistakes, the most impactful ones, sorted ascending by ply.
- Use UCI notation (e.g. e2e4, g1f3, e1g1) for "move" and "better" fields.
- "severity" must be exactly one of: inaccuracy, mistake, blunder.
- If there are no mistakes, return an empty array for "mistakes".
- Return ONLY the JSON object. No commentary outside the JSON.

Example output:
{"summary":"A well-played game with one critical error in the endgame.","mistakes":[{"ply":32,"move":"d4d5","severity":"blunder","better":"f2f4","comment":"Pushing the pawn allows Black to fork the rooks."}]}

PGN to analyse:
`
)

// Client calls the Gemini REST API to analyse chess games.
type Client struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// New returns a Client configured with the given API key and sensible defaults.
func New(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   defaultModel,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// WithModel returns a copy of the client targeting a different model id
// (e.g. "gemini-2.5-flash", "gemini-1.5-flash").
func (c *Client) WithModel(model string) *Client {
	cp := *c
	cp.model = model
	return &cp
}

// withBaseURL returns a copy of the client with a different base URL; used only in tests.
func (c *Client) withBaseURL(u string) *Client {
	cp := *c
	cp.baseURL = u
	return &cp
}

// geminiRequest mirrors the Gemini generateContent REST request body.
type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseMIMEType string  `json:"responseMimeType"`
	Temperature      float64 `json:"temperature"`
}

// geminiResponse is the subset of the Gemini REST response we care about.
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Analyze sends the PGN to Gemini and returns structured coaching feedback.
func (c *Client) Analyze(ctx context.Context, pgn string) (ports.Analysis, error) {
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: analyzePrompt + pgn}}},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			Temperature:      0.4,
		},
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: read body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := body
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return ports.Analysis{}, fmt.Errorf("gemini: status %d: %s", resp.StatusCode, snippet)
	}

	var gr geminiResponse
	if err = json.Unmarshal(body, &gr); err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: unmarshal response: %w", err)
	}

	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return ports.Analysis{}, fmt.Errorf("gemini: no candidates in response")
	}

	text := gr.Candidates[0].Content.Parts[0].Text
	text = stripCodeFence(text)

	var analysis ports.Analysis
	if err = json.Unmarshal([]byte(text), &analysis); err != nil {
		return ports.Analysis{}, fmt.Errorf("gemini: parse analysis JSON: %w", err)
	}

	normaliseSeverities(&analysis)
	return analysis, nil
}

// stripCodeFence removes optional ```json ... ``` markdown fences from s.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Drop the opening fence line
		idx := strings.Index(s, "\n")
		if idx != -1 {
			s = s[idx+1:]
		}
		// Drop the closing fence
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
		s = strings.TrimSpace(s)
	}
	return s
}

// normaliseSeverities ensures every Mistake has a valid Severity value.
// Unknown values are replaced with SevMistake.
func normaliseSeverities(a *ports.Analysis) {
	for i := range a.Mistakes {
		switch a.Mistakes[i].Severity {
		case ports.SevInaccuracy, ports.SevMistake, ports.SevBlunder:
			// valid, leave as-is
		default:
			a.Mistakes[i].Severity = ports.SevMistake
		}
	}
}
