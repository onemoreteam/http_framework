package httpframework

import (
	"net/http"
	"regexp"
	"strings"
)

// request matcher
type Matcher func(r *http.Request) bool

// match request by content-type
func MatchContentType(contentType string) Matcher {
	return func(r *http.Request) bool {
		return r.Header.Get("Content-Type") == contentType
	}
}

func grpcContentTypeMatcher(r *http.Request) bool {
	const ContentType = "application/grpc"
	return r.Header.Get("Content-Type") == ContentType || r.Header.Get("content-type") == ContentType
}

// match grpc request content-type
func MatchGrpcContentType() Matcher { return grpcContentTypeMatcher }

func grpcGatewayContentTypeMatcher(r *http.Request) bool {
	const ContentType = "application/grpc-gateway"
	return r.Header.Get("Content-Type") == ContentType || r.Header.Get("content-type") == ContentType
}

// match grpc-gateway request content-type
func MatchGrpcGatewayContentType() Matcher { return grpcGatewayContentTypeMatcher }

// match request by path prefix
func MatchPathPrefix(prefix string) Matcher {
	return func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, prefix)
	}
}

// match request by path regex
func MatchPathRegex(pattern string) Matcher {
	re := regexp.MustCompile(pattern)
	return func(r *http.Request) bool {
		return re.MatchString(r.URL.Path)
	}
}
