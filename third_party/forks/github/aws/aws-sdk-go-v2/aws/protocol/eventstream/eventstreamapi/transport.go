//go:build go1.18
// +build go1.18

package eventstreamapi

import smithyhttp "github.com/icha-senpai/note/third_party/forks/github/aws/smithy-go/transport/http"

// ApplyHTTPTransportFixes applies fixes to the HTTP request for proper event stream functionality.
//
// This operation is a no-op for Go 1.18 and above.
func ApplyHTTPTransportFixes(r *smithyhttp.Request) error {
	return nil
}
