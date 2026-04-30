package utils

import (
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
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

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string) string {
	replacer := strings.NewReplacer(
		"<script>", "",
		"</script>", "",
	)
	return replacer.Replace(input)
}
