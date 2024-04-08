// Package httpsmiddleware is used to ensure that the broker only accepts HTTPS connections
package httpsmiddleware

import (
	"net/http"
	"strings"

	"code.cloudfoundry.org/lager/v3"
)

func EnsureHTTPS(next http.Handler, logger lager.Logger, disable bool) http.Handler {
	if disable {
		logger.Info("https-redirect-disabled")
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto := r.Header.Values("X-Forwarded-Proto")
		if isHTTPS(proto) {
			next.ServeHTTP(w, r)
			return
		}

		location := "https://" + r.Host + r.URL.String()
		logger.Debug("redirecting-to-https", lager.Data{"to": location, "X-Forwarded-Proto": proto})
		http.Redirect(w, r, location, http.StatusMovedPermanently)
	})
}

// isHTTPS returns true when the protocol is definitely HTTPS
//
// From the docs: https://docs.vmware.com/en/VMware-Tanzu-Application-Service/4.0/tas-for-vms/http-routing.html
//
//	Developers can configure their apps to reject insecure requests by inspecting the X-Forwarded-Proto
//	HTTP header on incoming traffic. The header may have multiple values represented as a comma-separated
//	list, so developers must ensure the app rejects traffic that includes any X-Forwarded-Proto values
//	that are not HTTPS.
func isHTTPS(xForwardProto []string) bool {
	xForwardProto = strings.Split(strings.Join(xForwardProto, ","), ",")
	if len(xForwardProto) == 0 {
		return false
	}

	for _, proto := range xForwardProto {
		if proto != "https" {
			return false
		}
	}

	return true
}
