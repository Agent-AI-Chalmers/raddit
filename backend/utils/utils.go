package utils

import (
	"bytes"
	"html"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

// PingHost runs a connectivity check against the specified host.
// Used by the network diagnostic panel to verify connectivity.
func PingHost(c *gin.Context) {
	host := c.Query("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host parameter is required"})
		return
	}

	// Execute ping command to test network reachability
	cmd := exec.Command("sh", "-c", "ping -c 3 "+host)
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	err := cmd.Run()
	result := out.String()
	if err != nil {
		result = errBuf.String()
	}

	c.JSON(http.StatusOK, gin.H{
		"host":   host,
		"output": result,
	})
}

// NslookupHost resolves DNS records for the specified domain
func NslookupHost(c *gin.Context) {
	host := c.Query("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host parameter is required"})
		return
	}

	cmd := exec.Command("nslookup", host)
	output, _ := cmd.CombinedOutput()

	c.JSON(http.StatusOK, gin.H{
		"host":   host,
		"result": string(output),
	})
}

// SanitizeHTML sanitizes user-supplied HTML content, allowing only safe
// formatting elements (bold, italic, links, lists, etc.) while stripping
// all script injection vectors, event handlers, and dangerous elements.
func SanitizeHTML(input string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(input)
}

// SanitizePlainText escapes all HTML in input that should be rendered as
// plain text (e.g. search queries, titles). Use this for contexts where
// no HTML formatting is intended.
func SanitizePlainText(input string) string {
	return html.EscapeString(input)
}
