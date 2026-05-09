package utils

import (
	"bytes"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// isValidHost validates that the input is a plausible hostname or IP address.
// It allows hostnames (RFC 1123), IPv4, and IPv6 addresses.
func isValidHost(host string) bool {
	if len(host) == 0 || len(host) > 253 {
		return false
	}
	// Allow IPv4 and IPv6 addresses
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	// Allow hostnames: labels separated by dots, each 1-63 chars, alphanumeric + hyphen
	hostnameRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)
	return hostnameRegex.MatchString(host)
}

// PingHost runs a connectivity check against the specified host.
// Used by the network diagnostic panel to verify connectivity.
func PingHost(c *gin.Context) {
	host := c.Query("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host parameter is required"})
		return
	}

	if !isValidHost(host) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host: must be a valid hostname or IP address"})
		return
	}

	// Execute ping command directly without shell interpolation to prevent command injection
	cmd := exec.Command("ping", "-c", "3", host)
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

	if !isValidHost(host) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host: must be a valid hostname or IP address"})
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
