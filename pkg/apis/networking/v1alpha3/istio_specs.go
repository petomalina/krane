package v1alpha3

type VirtualServiceSpec struct {
	// REQUIRED. The destination hosts to which traffic is being sent. Could
	// be a DNS name with wildcard prefix or an IP address.  Depending on the
	// platform, short-names can also be used instead of a FQDN (i.e. has no
	// dots in the name). In such a scenario, the FQDN of the host would be
	// derived based on the underlying platform.
	//
	// A single VirtualService can be used to describe all the traffic
	// properties of the corresponding hosts, including those for multiple
	// HTTP and TCP ports. Alternatively, the traffic properties of a host
	// can be defined using more than one VirtualService, with certain
	// caveats. Refer to the
	// [Operations Guide](/help/ops/traffic-management/deploy-guidelines/#multiple-virtual-services-and-destination-rules-for-the-same-host)
	// for details.
	//
	// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
	// instead of "reviews.default.svc.cluster.local"), Istio will interpret
	// the short name based on the namespace of the rule, not the service. A
	// rule in the "default" namespace containing a host "reviews will be
	// interpreted as "reviews.default.svc.cluster.local", irrespective of
	// the actual namespace associated with the reviews service. _To avoid
	// potential misconfigurations, it is recommended to always use fully
	// qualified domain names over short names._
	//
	// The hosts field applies to both HTTP and TCP services. Service inside
	// the mesh, i.e., those found in the service registry, must always be
	// referred to using their alphanumeric names. IP addresses are allowed
	// only for services defined via the Gateway.
	Hosts []string `protobuf:"bytes,1,rep,name=hosts,proto3" json:"hosts,omitempty"`
	// The names of gateways that should apply these rules. If omitted, the
	// rules are only applied to sidecars inside the mesh. If a list of gateway
	// names is provided, the rules will not be applied to sidecars inside the
	// mesh unless the reserved gateway name `mesh` is included in the list.
	// Gateways defined in a different namespace can be selected by prefixing
	// the gateway name with `namespace/`.
	//
	// The selection condition imposed by this field can be overridden using
	// the source field in the match conditions of protocol-specific routes.
	Gateways []string `protobuf:"bytes,2,rep,name=gateways,proto3" json:"gateways,omitempty"`
	// An ordered list of route rules for HTTP traffic. HTTP routes will be
	// applied to platform service ports named 'http-*'/'http2-*'/'grpc-*', gateway
	// ports with protocol HTTP/HTTP2/GRPC/ TLS-terminated-HTTPS and service
	// entry ports using HTTP/HTTP2/GRPC protocols.  The first rule matching
	// an incoming request is used.
	Http []*HTTPRoute `protobuf:"bytes,3,rep,name=http,proto3" json:"http,omitempty"`
	// An ordered list of route rule for non-terminated TLS & HTTPS
	// traffic. Routing is typically performed using the SNI value presented
	// by the ClientHello message. TLS routes will be applied to platform
	// service ports named 'https-*', 'tls-*', unterminated gateway ports using
	// HTTPS/TLS protocols (i.e. with "passthrough" TLS mode) and service
	// entry ports using HTTPS/TLS protocols.  The first rule matching an
	// incoming request is used.  NOTE: Traffic 'https-*' or 'tls-*' ports
	// without associated virtual service will be treated as opaque TCP
	// traffic.
	Tls []*TLSRoute `protobuf:"bytes,5,rep,name=tls,proto3" json:"tls,omitempty"`
	// An ordered list of route rules for opaque TCP traffic. TCP routes will
	// be applied to any port that is not a HTTP or TLS port. The first rule
	// matching an incoming request is used.
	Tcp []*TCPRoute `protobuf:"bytes,4,rep,name=tcp,proto3" json:"tcp,omitempty"`
	// A list of namespaces to which this virtual service is exported. Exporting a
	// virtual service allows it to be used by sidecars and gateways defined in
	// other namespaces. This feature provides a mechanism for service owners
	// and mesh administrators to control the visibility of virtual services
	// across namespace boundaries.
	//
	// If no namespaces are specified then the virtual service is exported to all
	// namespaces by default.
	//
	// The value "." is reserved and defines an export to the same namespace that
	// the virtual service is declared in. Similarly the value "*" is reserved and
	// defines an export to all namespaces.
	//
	// NOTE: in the current release, the `exportTo` value is restricted to
	// "." or "*" (i.e., the current namespace or all namespaces).
	ExportTo []string `protobuf:"bytes,6,rep,name=export_to,json=exportTo,proto3" json:"export_to,omitempty"`
}

type Destination struct {
	// REQUIRED. The name of a service from the service registry. Service
	// names are looked up from the platform's service registry (e.g.,
	// Kubernetes services, Consul services, etc.) and from the hosts
	// declared by [ServiceEntry](/docs/reference/config/networking/v1alpha3/service-entry/#ServiceEntry). Traffic forwarded to
	// destinations that are not found in either of the two, will be dropped.
	//
	// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
	// instead of "reviews.default.svc.cluster.local"), Istio will interpret
	// the short name based on the namespace of the rule, not the service. A
	// rule in the "default" namespace containing a host "reviews will be
	// interpreted as "reviews.default.svc.cluster.local", irrespective of
	// the actual namespace associated with the reviews service. _To avoid
	// potential misconfigurations, it is recommended to always use fully
	// qualified domain names over short names._
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// The name of a subset within the service. Applicable only to services
	// within the mesh. The subset must be defined in a corresponding
	// DestinationRule.
	Subset string `protobuf:"bytes,2,opt,name=subset,proto3" json:"subset,omitempty"`
	// Specifies the port on the host that is being addressed. If a service
	// exposes only a single port it is not required to explicitly select the
	// port.
	Port *PortSelector `protobuf:"bytes,3,opt,name=port,proto3" json:"port,omitempty"`
}

// Describes match conditions and actions for routing HTTP/1.1, HTTP2, and
// gRPC traffic. See VirtualService for usage examples.
type HTTPRoute struct {
	// Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*HTTPMatchRequest `protobuf:"bytes,1,rep,name=match,proto3" json:"match,omitempty"`
	// A http rule can either redirect or forward (default) traffic. The
	// forwarding target can be one of several versions of a service (see
	// glossary in beginning of document). Weights associated with the
	// service version determine the proportion of traffic it receives.
	Route []*HTTPRouteDestination `protobuf:"bytes,2,rep,name=route,proto3" json:"route,omitempty"`
	// A http rule can either redirect or forward (default) traffic. If
	// traffic passthrough option is specified in the rule,
	// route/redirect will be ignored. The redirect primitive can be used to
	// send a HTTP 301 redirect to a different URI or Authority.
	Redirect *HTTPRedirect `protobuf:"bytes,3,opt,name=redirect,proto3" json:"redirect,omitempty"`
	// Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with
	// Redirect primitive. Rewrite will be performed before forwarding.
	Rewrite *HTTPRewrite `protobuf:"bytes,4,opt,name=rewrite,proto3" json:"rewrite,omitempty"`
	// Deprecated. Websocket upgrades are done automatically starting from Istio 1.0.
	// $hide_from_docs
	WebsocketUpgrade bool `protobuf:"varint,5,opt,name=websocket_upgrade,json=websocketUpgrade,proto3" json:"websocket_upgrade,omitempty"`
	// Timeout for HTTP requests.
	Timeout string `protobuf:"bytes,6,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// Retry policy for HTTP requests.
	Retries *HTTPRetry `protobuf:"bytes,7,opt,name=retries,proto3" json:"retries,omitempty"`
	// Fault injection policy to apply on HTTP traffic at the client side.
	// Note that timeouts or retries will not be enabled when faults are
	// enabled on the client side.
	Fault *HTTPFaultInjection `protobuf:"bytes,8,opt,name=fault,proto3" json:"fault,omitempty"`
	// Mirror HTTP traffic to a another destination in addition to forwarding
	// the requests to the intended destination. Mirrored traffic is on a
	// best effort basis where the sidecar/gateway will not wait for the
	// mirrored cluster to respond before returning the response from the
	// original destination.  Statistics will be generated for the mirrored
	// destination.
	Mirror *Destination `protobuf:"bytes,9,opt,name=mirror,proto3" json:"mirror,omitempty"`
	// Cross-Origin Resource Sharing policy (CORS). Refer to
	// [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
	// for further details about cross origin resource sharing.
	CorsPolicy *CorsPolicy `protobuf:"bytes,10,opt,name=cors_policy,json=corsPolicy,proto3" json:"cors_policy,omitempty"`
	// Use of `append_headers` is deprecated. Use the `headers`
	// field instead.
	AppendHeaders map[string]string `protobuf:"bytes,11,rep,name=append_headers,json=appendHeaders,proto3" json:"append_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// Use of `remove_response_header` is deprecated. Use the `headers`
	// field instead.
	RemoveResponseHeaders []string `protobuf:"bytes,12,rep,name=remove_response_headers,json=removeResponseHeaders,proto3" json:"remove_response_headers,omitempty"` // Deprecated: Do not use.
	// Use of `append_response_headers` is deprecated. Use the `headers`
	// field instead.
	AppendResponseHeaders map[string]string `protobuf:"bytes,13,rep,name=append_response_headers,json=appendResponseHeaders,proto3" json:"append_response_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// Use of `remove_request_headers` is deprecated. Use the `headers`
	// field instead.
	RemoveRequestHeaders []string `protobuf:"bytes,14,rep,name=remove_request_headers,json=removeRequestHeaders,proto3" json:"remove_request_headers,omitempty"` // Deprecated: Do not use.
	// Use of `append_request_headers` is deprecated. Use the `headers`
	// field instead.
	AppendRequestHeaders map[string]string `protobuf:"bytes,15,rep,name=append_request_headers,json=appendRequestHeaders,proto3" json:"append_request_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// Header manipulation rules
	Headers *Headers `protobuf:"bytes,16,opt,name=headers,proto3" json:"headers,omitempty"`
}

// Header manipulation rules
type Headers struct {
	// Header manipulation rules to apply before forwarding a request
	// to the destination service
	Request *Headers_HeaderOperations `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	// Header manipulation rules to apply before returning a response
	// to the caller
	Response *Headers_HeaderOperations `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
}

// HeaderOperations Describes the header manipulations to apply
type Headers_HeaderOperations struct {
	// Overwrite the headers specified by key with the given values
	Set map[string]string `protobuf:"bytes,1,rep,name=set,proto3" json:"set,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Append the given values to the headers specified by keys
	// (will create a comma-separated list of values)
	Add map[string]string `protobuf:"bytes,2,rep,name=add,proto3" json:"add,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Remove a the specified headers
	Remove []string `protobuf:"bytes,3,rep,name=remove,proto3" json:"remove,omitempty"`
}

type TLSRoute struct {
	// REQUIRED. Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*TLSMatchAttributes `protobuf:"bytes,1,rep,name=match,proto3" json:"match,omitempty"`
	// The destination to which the connection should be forwarded to.
	Route []*RouteDestination `protobuf:"bytes,2,rep,name=route,proto3" json:"route,omitempty"`
}

type TCPRoute struct {
	// Match conditions to be satisfied for the rule to be
	// activated. All conditions inside a single match block have AND
	// semantics, while the list of match blocks have OR semantics. The rule
	// is matched if any one of the match blocks succeed.
	Match []*L4MatchAttributes `protobuf:"bytes,1,rep,name=match,proto3" json:"match,omitempty"`
	// The destination to which the connection should be forwarded to.
	Route []*RouteDestination `protobuf:"bytes,2,rep,name=route,proto3" json:"route,omitempty"`
}

type HTTPMatchRequest struct {
	// URI to match
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Uri *StringMatch `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// URI Scheme
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Scheme *StringMatch `protobuf:"bytes,2,opt,name=scheme,proto3" json:"scheme,omitempty"`
	// HTTP Method
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Method *StringMatch `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	// HTTP Authority
	// values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	Authority *StringMatch `protobuf:"bytes,4,opt,name=authority,proto3" json:"authority,omitempty"`
	// The header keys must be lowercase and use hyphen as the separator,
	// e.g. _x-request-id_.
	//
	// Header values are case-sensitive and formatted as follows:
	//
	// - `exact: "value"` for exact string match
	//
	// - `prefix: "value"` for prefix-based match
	//
	// - `regex: "value"` for ECMAscript style regex-based match
	//
	// **Note:** The keys `uri`, `scheme`, `method`, and `authority` will be ignored.
	Headers map[string]*StringMatch `protobuf:"bytes,5,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Specifies the ports on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they support,
	// in these cases it is not required to explicitly select the port.
	Port uint32 `protobuf:"varint,6,opt,name=port,proto3" json:"port,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,7,rep,name=source_labels,json=sourceLabels,proto3" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway match is
	// independent of sourceLabels.
	Gateways []string `protobuf:"bytes,8,rep,name=gateways,proto3" json:"gateways,omitempty"`
}

// Each routing rule is associated with one or more service versions (see
// glossary in beginning of document). Weights associated with the version
// determine the proportion of traffic it receives. For example, the
// following rule will route 25% of traffic for the "reviews" service to
// instances with the "v2" tag and the remaining traffic (i.e., 75%) to
// "v1".
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: reviews-route
// spec:
//   hosts:
//   - reviews.prod.svc.cluster.local
//   http:
//   - route:
//     - destination:
//         host: reviews.prod.svc.cluster.local
//         subset: v2
//       weight: 25
//     - destination:
//         host: reviews.prod.svc.cluster.local
//         subset: v1
//       weight: 75
// ```
//
// And the associated DestinationRule
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: DestinationRule
// metadata:
//   name: reviews-destination
// spec:
//   host: reviews.prod.svc.cluster.local
//   subsets:
//   - name: v1
//     labels:
//       version: v1
//   - name: v2
//     labels:
//       version: v2
// ```
//
// Traffic can also be split across two entirely different services without
// having to define new subsets. For example, the following rule forwards 25% of
// traffic to reviews.com to dev.reviews.com
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: reviews-route-two-domains
// spec:
//   hosts:
//   - reviews.com
//   http:
//   - route:
//     - destination:
//         host: dev.reviews.com
//       weight: 25
//     - destination:
//         host: reviews.com
//       weight: 75
// ```
type HTTPRouteDestination struct {
	// REQUIRED. Destination uniquely identifies the instances of a service
	// to which the request/connection should be forwarded to.
	Destination *Destination `protobuf:"bytes,1,opt,name=destination,proto3" json:"destination,omitempty"`
	// REQUIRED. The proportion of traffic to be forwarded to the service
	// version. (0-100). Sum of weights across destinations SHOULD BE == 100.
	// If there is only one destination in a rule, the weight value is assumed to
	// be 100.
	Weight int32 `protobuf:"varint,2,opt,name=weight,proto3" json:"weight,omitempty"`
	// Use of `remove_response_header` is deprecated. Use the `headers`
	// field instead.
	RemoveResponseHeaders []string `protobuf:"bytes,3,rep,name=remove_response_headers,json=removeResponseHeaders,proto3" json:"remove_response_headers,omitempty"` // Deprecated: Do not use.
	// Use of `append_response_headers` is deprecated. Use the `headers`
	// field instead.
	AppendResponseHeaders map[string]string `protobuf:"bytes,4,rep,name=append_response_headers,json=appendResponseHeaders,proto3" json:"append_response_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// Use of `remove_request_headers` is deprecated. Use the `headers`
	// field instead.
	RemoveRequestHeaders []string `protobuf:"bytes,5,rep,name=remove_request_headers,json=removeRequestHeaders,proto3" json:"remove_request_headers,omitempty"` // Deprecated: Do not use.
	// Use of `append_request_headers` is deprecated. Use the `headers`
	// field instead.
	AppendRequestHeaders map[string]string `protobuf:"bytes,6,rep,name=append_request_headers,json=appendRequestHeaders,proto3" json:"append_request_headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // Deprecated: Do not use.
	// Header manipulation rules
	Headers *Headers `protobuf:"bytes,7,opt,name=headers,proto3" json:"headers,omitempty"`
}

// L4 routing rule weighted destination.
type RouteDestination struct {
	// REQUIRED. Destination uniquely identifies the instances of a service
	// to which the request/connection should be forwarded to.
	Destination *Destination `protobuf:"bytes,1,opt,name=destination,proto3" json:"destination,omitempty"`
	// REQUIRED. The proportion of traffic to be forwarded to the service
	// version. If there is only one destination in a rule, all traffic will be
	// routed to it irrespective of the weight.
	Weight int32 `protobuf:"varint,2,opt,name=weight,proto3" json:"weight,omitempty"`
}

func (m *RouteDestination) GetDestination() *Destination {
	if m != nil {
		return m.Destination
	}
	return nil
}

func (m *RouteDestination) GetWeight() int32 {
	if m != nil {
		return m.Weight
	}
	return 0
}

// L4 connection match attributes. Note that L4 connection matching support
// is incomplete.
type L4MatchAttributes struct {
	// IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g.,
	// a.b.c.d/xx form or just a.b.c.d.
	DestinationSubnets []string `protobuf:"bytes,1,rep,name=destination_subnets,json=destinationSubnets,proto3" json:"destination_subnets,omitempty"`
	// Specifies the port on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they support,
	// in these cases it is not required to explicitly select the port.
	Port uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx
	// form or just a.b.c.d
	// $hide_from_docs
	SourceSubnet string `protobuf:"bytes,3,opt,name=source_subnet,json=sourceSubnet,proto3" json:"source_subnet,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,4,rep,name=source_labels,json=sourceLabels,proto3" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway
	// match is independent of sourceLabels.
	Gateways []string `protobuf:"bytes,5,rep,name=gateways,proto3" json:"gateways,omitempty"`
}

// TLS connection match attributes.
type TLSMatchAttributes struct {
	// REQUIRED. SNI (server name indicator) to match on. Wildcard prefixes
	// can be used in the SNI value, e.g., *.com will match foo.example.com
	// as well as example.com. An SNI value must be a subset (i.e., fall
	// within the domain) of the corresponding virtual serivce's hosts.
	SniHosts []string `protobuf:"bytes,1,rep,name=sni_hosts,json=sniHosts,proto3" json:"sni_hosts,omitempty"`
	// IPv4 or IPv6 ip addresses of destination with optional subnet.  E.g.,
	// a.b.c.d/xx form or just a.b.c.d.
	DestinationSubnets []string `protobuf:"bytes,2,rep,name=destination_subnets,json=destinationSubnets,proto3" json:"destination_subnets,omitempty"`
	// Specifies the port on the host that is being addressed. Many services
	// only expose a single port or label ports with the protocols they
	// support, in these cases it is not required to explicitly select the
	// port.
	Port uint32 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	// IPv4 or IPv6 ip address of source with optional subnet. E.g., a.b.c.d/xx
	// form or just a.b.c.d
	// $hide_from_docs
	SourceSubnet string `protobuf:"bytes,4,opt,name=source_subnet,json=sourceSubnet,proto3" json:"source_subnet,omitempty"`
	// One or more labels that constrain the applicability of a rule to
	// workloads with the given labels. If the VirtualService has a list of
	// gateways specified at the top, it should include the reserved gateway
	// `mesh` in order for this field to be applicable.
	SourceLabels map[string]string `protobuf:"bytes,5,rep,name=source_labels,json=sourceLabels,proto3" json:"source_labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Names of gateways where the rule should be applied to. Gateway names
	// at the top of the VirtualService (if any) are overridden. The gateway
	// match is independent of sourceLabels.
	Gateways []string `protobuf:"bytes,6,rep,name=gateways,proto3" json:"gateways,omitempty"`
}

// HTTPRedirect can be used to send a 301 redirect response to the caller,
// where the Authority/Host and the URI in the response can be swapped with
// the specified values. For example, the following rule redirects
// requests for /v1/getProductRatings API on the ratings service to
// /v1/bookRatings provided by the bookratings service.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: ratings-route
// spec:
//   hosts:
//   - ratings.prod.svc.cluster.local
//   http:
//   - match:
//     - uri:
//         exact: /v1/getProductRatings
//     redirect:
//       uri: /v1/bookRatings
//       authority: newratings.default.svc.cluster.local
//   ...
// ```
type HTTPRedirect struct {
	// On a redirect, overwrite the Path portion of the URL with this
	// value. Note that the entire path will be replaced, irrespective of the
	// request URI being matched as an exact path or prefix.
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// On a redirect, overwrite the Authority/Host portion of the URL with
	// this value.
	Authority string `protobuf:"bytes,2,opt,name=authority,proto3" json:"authority,omitempty"`
}

// HTTPRewrite can be used to rewrite specific parts of a HTTP request
// before forwarding the request to the destination. Rewrite primitive can
// be used only with HTTPRouteDestination. The following example
// demonstrates how to rewrite the URL prefix for api call (/ratings) to
// ratings service before making the actual API call.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: ratings-route
// spec:
//   hosts:
//   - ratings.prod.svc.cluster.local
//   http:
//   - match:
//     - uri:
//         prefix: /ratings
//     rewrite:
//       uri: /v1/bookRatings
//     route:
//     - destination:
//         host: ratings.prod.svc.cluster.local
//         subset: v1
// ```
//
type HTTPRewrite struct {
	// rewrite the path (or the prefix) portion of the URI with this
	// value. If the original URI was matched based on prefix, the value
	// provided in this field will replace the corresponding matched prefix.
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	// rewrite the Authority/Host header with this value.
	Authority string `protobuf:"bytes,2,opt,name=authority,proto3" json:"authority,omitempty"`
}

// Describes how to match a given string in HTTP headers. Match is
// case-sensitive.
type StringMatch struct {
	// Types that are valid to be assigned to MatchType:
	//	*StringMatch_Exact
	//	*StringMatch_Prefix
	//	*StringMatch_Regex
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"regex,omitempty"`
	Regex  string `json:"regex,omitempty"`
}

type HTTPRetry struct {
	// REQUIRED. Number of retries for a given request. The interval
	// between retries will be determined automatically (25ms+). Actual
	// number of retries attempted depends on the httpReqTimeout.
	Attempts int32 `protobuf:"varint,1,opt,name=attempts,proto3" json:"attempts,omitempty"`
	// Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms.
	PerTryTimeout string `protobuf:"bytes,2,opt,name=per_try_timeout,json=perTryTimeout,proto3" json:"per_try_timeout,omitempty"`
	// Specifies the conditions under which retry takes place.
	// One or more policies can be specified using a ‘,’ delimited list.
	// See the [supported policies](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on)
	// and [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-grpc-on) for more details.
	RetryOn string `protobuf:"bytes,3,opt,name=retry_on,json=retryOn,proto3" json:"retry_on,omitempty"`
}

type CorsPolicy struct {
	// The list of origins that are allowed to perform CORS requests. The
	// content will be serialized into the Access-Control-Allow-Origin
	// header. Wildcard * will allow all origins.
	AllowOrigin []string `protobuf:"bytes,1,rep,name=allow_origin,json=allowOrigin,proto3" json:"allow_origin,omitempty"`
	// List of HTTP methods allowed to access the resource. The content will
	// be serialized into the Access-Control-Allow-Methods header.
	AllowMethods []string `protobuf:"bytes,2,rep,name=allow_methods,json=allowMethods,proto3" json:"allow_methods,omitempty"`
	// List of HTTP headers that can be used when requesting the
	// resource. Serialized to Access-Control-Allow-Headers header.
	AllowHeaders []string `protobuf:"bytes,3,rep,name=allow_headers,json=allowHeaders,proto3" json:"allow_headers,omitempty"`
	// A white list of HTTP headers that the browsers are allowed to
	// access. Serialized into Access-Control-Expose-Headers header.
	ExposeHeaders []string `protobuf:"bytes,4,rep,name=expose_headers,json=exposeHeaders,proto3" json:"expose_headers,omitempty"`
	// Specifies how long the results of a preflight request can be
	// cached. Translates to the `Access-Control-Max-Age` header.
	MaxAge string `protobuf:"bytes,5,opt,name=max_age,json=maxAge,proto3" json:"max_age,omitempty"`
	// Indicates whether the caller is allowed to send the actual request
	// (not the preflight) using credentials. Translates to
	// `Access-Control-Allow-Credentials` header.
	AllowCredentials bool `protobuf:"bytes,6,opt,name=allow_credentials,json=allowCredentials,proto3" json:"allow_credentials,omitempty"`
}

type HTTPFaultInjection struct {
	// Delay requests before forwarding, emulating various failures such as
	// network issues, overloaded upstream service, etc.
	Delay *HTTPFaultInjection_Delay `protobuf:"bytes,1,opt,name=delay,proto3" json:"delay,omitempty"`
	// Abort Http request attempts and return error codes back to downstream
	// service, giving the impression that the upstream service is faulty.
	Abort *HTTPFaultInjection_Abort `protobuf:"bytes,2,opt,name=abort,proto3" json:"abort,omitempty"`
}

type HTTPFaultInjection_Delay struct {
	// Percentage of requests on which the delay will be injected (0-100).
	// Use of integer `percent` value is deprecated. Use the double `percentage`
	// field instead.
	Percent int32 `protobuf:"varint,1,opt,name=percent,proto3" json:"percent,omitempty"` // Deprecated: Do not use.
	// Types that are valid to be assigned to HttpDelayType:
	//	*HTTPFaultInjection_Delay_FixedDelay
	//	*HTTPFaultInjection_Delay_ExponentialDelay

	FixedDelay       string `json:"fixedDelay,omitempty"`
	ExponentialDelay string `json:"exponentialDelay,omitempty"`

	// Percentage of requests on which the delay will be injected.
	Percentage *Percent `protobuf:"bytes,5,opt,name=percentage,proto3" json:"percentage,omitempty"`
}

type HTTPFaultInjection_Abort struct {
	// Percentage of requests to be aborted with the error code provided (0-100).
	// Use of integer `percent` value is deprecated. Use the double `percentage`
	// field instead.
	Percent int32 `protobuf:"varint,1,opt,name=percent,proto3" json:"percent,omitempty"` // Deprecated: Do not use.
	// Types that are valid to be assigned to ErrorType:
	//	*HTTPFaultInjection_Abort_HttpStatus
	//	*HTTPFaultInjection_Abort_GrpcStatus
	//	*HTTPFaultInjection_Abort_Http2Error
	HttpStatus int32  `json:"httpStatus,omitempty"`
	GrpcStatus string `json:"grpcStatus,omitempty"`
	Http2Error string `json:"http2Error,omitempty"`
	// Percentage of requests to be aborted with the error code provided.
	Percentage *Percent `protobuf:"bytes,5,opt,name=percentage,proto3" json:"percentage,omitempty"`
}

// PortSelector specifies the number of a port to be used for
// matching or selection for final routing.
type PortSelector struct {
	Number uint32 `json:"number,omitempty"`
	Name   string `json:"name,omitempty"`
}

// Percent specifies a percentage in the range of [0.0, 100.0].
type Percent struct {
	Value float64 `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
}
