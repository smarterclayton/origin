// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/filter/network/tcp_proxy.proto

package network

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import envoy_api_v2_filter1 "api/filter"
import envoy_api_v2 "api"
import google_protobuf1 "github.com/golang/protobuf/ptypes/duration"
import google_protobuf "github.com/golang/protobuf/ptypes/wrappers"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// [V2-API-DIFF] The route match now takes place in the FilterChainMatch table.
type TcpProxy struct {
	// The human readable prefix to use when emitting statistics for the
	// TCP proxy filter. See the statistics documentation for more information.
	StatPrefix string `protobuf:"bytes,1,opt,name=stat_prefix,json=statPrefix" json:"stat_prefix,omitempty"`
	// The upstream cluster to connect to.
	// TODO(htuch): look into shared WeightedCluster support with RDS.
	Cluster string `protobuf:"bytes,2,opt,name=cluster" json:"cluster,omitempty"`
	// [V2-API-DIFF] The idle timeout for connections managed by the TCP proxy
	// filter. The idle timeout is defined as the period in which there is no
	// active traffic. If not set, there is no idle timeout. When the idle timeout
	// is reached the connection will be closed. The distinction between
	// downstream_idle_timeout/upstream_idle_timeout provides a means to set
	// timeout based on the last byte sent on the downstream/upstream connection.
	DownstreamIdleTimeout *google_protobuf1.Duration `protobuf:"bytes,3,opt,name=downstream_idle_timeout,json=downstreamIdleTimeout" json:"downstream_idle_timeout,omitempty"`
	UpstreamIdleTimeout   *google_protobuf1.Duration `protobuf:"bytes,4,opt,name=upstream_idle_timeout,json=upstreamIdleTimeout" json:"upstream_idle_timeout,omitempty"`
	// Configuration for access logs.
	AccessLog    []*envoy_api_v2_filter1.AccessLog `protobuf:"bytes,5,rep,name=access_log,json=accessLog" json:"access_log,omitempty"`
	DeprecatedV1 *TcpProxy_DeprecatedV1            `protobuf:"bytes,6,opt,name=deprecated_v1,json=deprecatedV1" json:"deprecated_v1,omitempty"`
	// The maximum number of unsuccessful connection attempts that will be made before
	// giving up.  If the parameter is not specified, 1 connection attempt will be made.
	MaxConnectAttempts *google_protobuf.UInt32Value `protobuf:"bytes,7,opt,name=max_connect_attempts,json=maxConnectAttempts" json:"max_connect_attempts,omitempty"`
}

func (m *TcpProxy) Reset()                    { *m = TcpProxy{} }
func (m *TcpProxy) String() string            { return proto.CompactTextString(m) }
func (*TcpProxy) ProtoMessage()               {}
func (*TcpProxy) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *TcpProxy) GetStatPrefix() string {
	if m != nil {
		return m.StatPrefix
	}
	return ""
}

func (m *TcpProxy) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *TcpProxy) GetDownstreamIdleTimeout() *google_protobuf1.Duration {
	if m != nil {
		return m.DownstreamIdleTimeout
	}
	return nil
}

func (m *TcpProxy) GetUpstreamIdleTimeout() *google_protobuf1.Duration {
	if m != nil {
		return m.UpstreamIdleTimeout
	}
	return nil
}

func (m *TcpProxy) GetAccessLog() []*envoy_api_v2_filter1.AccessLog {
	if m != nil {
		return m.AccessLog
	}
	return nil
}

func (m *TcpProxy) GetDeprecatedV1() *TcpProxy_DeprecatedV1 {
	if m != nil {
		return m.DeprecatedV1
	}
	return nil
}

func (m *TcpProxy) GetMaxConnectAttempts() *google_protobuf.UInt32Value {
	if m != nil {
		return m.MaxConnectAttempts
	}
	return nil
}

type TcpProxy_DeprecatedV1 struct {
	// The route table for the filter. All filter instances must have a route
	// table, even if it is empty.
	Routes []*TcpProxy_DeprecatedV1_TCPRoute `protobuf:"bytes,1,rep,name=routes" json:"routes,omitempty"`
}

func (m *TcpProxy_DeprecatedV1) Reset()                    { *m = TcpProxy_DeprecatedV1{} }
func (m *TcpProxy_DeprecatedV1) String() string            { return proto.CompactTextString(m) }
func (*TcpProxy_DeprecatedV1) ProtoMessage()               {}
func (*TcpProxy_DeprecatedV1) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0, 0} }

func (m *TcpProxy_DeprecatedV1) GetRoutes() []*TcpProxy_DeprecatedV1_TCPRoute {
	if m != nil {
		return m.Routes
	}
	return nil
}

// [V2-API-DIFF] This is deprecated in v2. Routes will be matched using
// the FilterChainMatch in Listeners.
//
// A TCP proxy route consists of a set of optional L4 criteria and the
// name of a cluster. If a downstream connection matches all the
// specified criteria, the cluster in the route is used for the
// corresponding upstream connection. Routes are tried in the order
// specified until a match is found. If no match is found, the connection
// is closed. A route with no criteria is valid and always produces a
// match.
type TcpProxy_DeprecatedV1_TCPRoute struct {
	// The cluster to connect to when a the downstream network connection
	// matches the specified criteria.
	Cluster string `protobuf:"bytes,1,opt,name=cluster" json:"cluster,omitempty"`
	// An optional list of IP address subnets in the form
	// “ip_address/xx”. The criteria is satisfied if the destination IP
	// address of the downstream connection is contained in at least one of
	// the specified subnets. If the parameter is not specified or the list
	// is empty, the destination IP address is ignored. The destination IP
	// address of the downstream connection might be different from the
	// addresses on which the proxy is listening if the connection has been
	// redirected.
	DestinationIpList []*envoy_api_v2.CidrRange `protobuf:"bytes,2,rep,name=destination_ip_list,json=destinationIpList" json:"destination_ip_list,omitempty"`
	// An optional string containing a comma-separated list of port numbers
	// or ranges. The criteria is satisfied if the destination port of the
	// downstream connection is contained in at least one of the specified
	// ranges. If the parameter is not specified, the destination port is
	// ignored. The destination port address of the downstream connection
	// might be different from the port on which the proxy is listening if
	// the connection has been redirected.
	DestinationPorts string `protobuf:"bytes,3,opt,name=destination_ports,json=destinationPorts" json:"destination_ports,omitempty"`
	// An optional list of IP address subnets in the form
	// “ip_address/xx”. The criteria is satisfied if the source IP address
	// of the downstream connection is contained in at least one of the
	// specified subnets. If the parameter is not specified or the list is
	// empty, the source IP address is ignored.
	SourceIpList []*envoy_api_v2.CidrRange `protobuf:"bytes,4,rep,name=source_ip_list,json=sourceIpList" json:"source_ip_list,omitempty"`
	// An optional string containing a comma-separated list of port numbers
	// or ranges. The criteria is satisfied if the source port of the
	// downstream connection is contained in at least one of the specified
	// ranges. If the parameter is not specified, the source port is
	// ignored.
	SourcePorts string `protobuf:"bytes,5,opt,name=source_ports,json=sourcePorts" json:"source_ports,omitempty"`
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) Reset()         { *m = TcpProxy_DeprecatedV1_TCPRoute{} }
func (m *TcpProxy_DeprecatedV1_TCPRoute) String() string { return proto.CompactTextString(m) }
func (*TcpProxy_DeprecatedV1_TCPRoute) ProtoMessage()    {}
func (*TcpProxy_DeprecatedV1_TCPRoute) Descriptor() ([]byte, []int) {
	return fileDescriptor4, []int{0, 0, 0}
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) GetCluster() string {
	if m != nil {
		return m.Cluster
	}
	return ""
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) GetDestinationIpList() []*envoy_api_v2.CidrRange {
	if m != nil {
		return m.DestinationIpList
	}
	return nil
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) GetDestinationPorts() string {
	if m != nil {
		return m.DestinationPorts
	}
	return ""
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) GetSourceIpList() []*envoy_api_v2.CidrRange {
	if m != nil {
		return m.SourceIpList
	}
	return nil
}

func (m *TcpProxy_DeprecatedV1_TCPRoute) GetSourcePorts() string {
	if m != nil {
		return m.SourcePorts
	}
	return ""
}

func init() {
	proto.RegisterType((*TcpProxy)(nil), "envoy.api.v2.filter.network.TcpProxy")
	proto.RegisterType((*TcpProxy_DeprecatedV1)(nil), "envoy.api.v2.filter.network.TcpProxy.DeprecatedV1")
	proto.RegisterType((*TcpProxy_DeprecatedV1_TCPRoute)(nil), "envoy.api.v2.filter.network.TcpProxy.DeprecatedV1.TCPRoute")
}

func init() { proto.RegisterFile("api/filter/network/tcp_proxy.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 507 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x5f, 0x6f, 0xd3, 0x30,
	0x10, 0x57, 0xba, 0xad, 0x5d, 0xdd, 0x82, 0x98, 0xc7, 0xb4, 0x50, 0xd0, 0x28, 0x7b, 0xaa, 0x84,
	0xe4, 0x6a, 0xdd, 0x23, 0xda, 0xc3, 0xe8, 0x24, 0x54, 0x69, 0xa0, 0x12, 0xca, 0x90, 0x78, 0xb1,
	0xbc, 0xf8, 0x1a, 0x59, 0xa4, 0xb1, 0x65, 0x5f, 0xda, 0xee, 0x7b, 0xf1, 0x89, 0xf8, 0x1c, 0x3c,
	0xa0, 0x3a, 0xc9, 0xc8, 0xc4, 0x18, 0xe2, 0x2d, 0xf7, 0xbb, 0xfb, 0xfd, 0x89, 0xef, 0xc8, 0xb1,
	0x30, 0x6a, 0x38, 0x57, 0x29, 0x82, 0x1d, 0x66, 0x80, 0x2b, 0x6d, 0xbf, 0x0d, 0x31, 0x36, 0xdc,
	0x58, 0xbd, 0xbe, 0x61, 0xc6, 0x6a, 0xd4, 0xf4, 0x39, 0x64, 0x4b, 0x7d, 0xc3, 0x84, 0x51, 0x6c,
	0x39, 0x62, 0xc5, 0x30, 0x2b, 0x87, 0x7b, 0xbd, 0x9a, 0x80, 0x88, 0x63, 0x70, 0x2e, 0xd5, 0x49,
	0x41, 0xec, 0xed, 0x6d, 0x7a, 0x42, 0x4a, 0x0b, 0xce, 0x95, 0xd0, 0x51, 0xa2, 0x75, 0x92, 0xc2,
	0xd0, 0x57, 0xd7, 0xf9, 0x7c, 0x28, 0x73, 0x2b, 0x50, 0xe9, 0xec, 0x6f, 0xfd, 0x95, 0x15, 0xc6,
	0x80, 0x2d, 0xf9, 0xc7, 0xdf, 0x9b, 0x64, 0x77, 0x16, 0x9b, 0xe9, 0x26, 0x1e, 0x7d, 0x49, 0x3a,
	0x0e, 0x05, 0x72, 0x63, 0x61, 0xae, 0xd6, 0x61, 0xd0, 0x0f, 0x06, 0xed, 0x88, 0x6c, 0xa0, 0xa9,
	0x47, 0x68, 0x48, 0x5a, 0x71, 0x9a, 0x3b, 0x04, 0x1b, 0x36, 0x7c, 0xb3, 0x2a, 0xe9, 0x47, 0x72,
	0x28, 0xf5, 0x2a, 0x73, 0x68, 0x41, 0x2c, 0xb8, 0x92, 0x29, 0x70, 0x54, 0x0b, 0xd0, 0x39, 0x86,
	0x5b, 0xfd, 0x60, 0xd0, 0x19, 0x3d, 0x63, 0x45, 0x12, 0x56, 0x25, 0x61, 0x17, 0x65, 0xd2, 0xe8,
	0xe0, 0x37, 0x73, 0x22, 0x53, 0x98, 0x15, 0x3c, 0xfa, 0x9e, 0x1c, 0xe4, 0xe6, 0x3e, 0xc1, 0xed,
	0x7f, 0x09, 0xee, 0x57, 0xbc, 0xba, 0xdc, 0x19, 0x21, 0xc5, 0x7b, 0xf2, 0x54, 0x27, 0xe1, 0x4e,
	0x7f, 0x6b, 0xd0, 0x19, 0x1d, 0xb1, 0xfb, 0x56, 0x71, 0xee, 0xc7, 0x2e, 0x75, 0x12, 0xb5, 0x45,
	0xf5, 0x49, 0xbf, 0x90, 0x47, 0x12, 0x8c, 0x85, 0x58, 0x20, 0x48, 0xbe, 0x3c, 0x09, 0x9b, 0x3e,
	0xc5, 0x88, 0x3d, 0xb0, 0x4c, 0x56, 0xbd, 0x2c, 0xbb, 0xb8, 0xa5, 0x5e, 0x9d, 0x44, 0x5d, 0x59,
	0xab, 0xe8, 0x07, 0xf2, 0x74, 0x21, 0xd6, 0x3c, 0xd6, 0x59, 0x06, 0x31, 0x72, 0x81, 0x08, 0x0b,
	0x83, 0x2e, 0x6c, 0x79, 0xfd, 0x17, 0x7f, 0xfc, 0xe5, 0xe7, 0x49, 0x86, 0xa7, 0xa3, 0x2b, 0x91,
	0xe6, 0x10, 0xd1, 0x85, 0x58, 0x8f, 0x0b, 0xe2, 0x79, 0xc9, 0xeb, 0xfd, 0x68, 0x90, 0x6e, 0xdd,
	0x8e, 0x7e, 0x22, 0x4d, 0xab, 0x73, 0x04, 0x17, 0x06, 0xfe, 0xa7, 0xdf, 0xfc, 0x7f, 0x64, 0x36,
	0x1b, 0x4f, 0xa3, 0x8d, 0x46, 0x54, 0x4a, 0xf5, 0x7e, 0x06, 0x64, 0xb7, 0x02, 0xeb, 0x67, 0x11,
	0xdc, 0x3d, 0x8b, 0x77, 0x64, 0x5f, 0x82, 0x43, 0x95, 0xf9, 0xc5, 0x70, 0x65, 0x78, 0xaa, 0x1c,
	0x86, 0x0d, 0x1f, 0xe4, 0xf0, 0x6e, 0x90, 0xb1, 0x92, 0x36, 0x12, 0x59, 0x02, 0xd1, 0x5e, 0x8d,
	0x33, 0x31, 0x97, 0xca, 0x21, 0x7d, 0x4d, 0xea, 0x20, 0x37, 0xda, 0xa2, 0xf3, 0x97, 0xd5, 0x8e,
	0x9e, 0xd4, 0x1a, 0xd3, 0x0d, 0x4e, 0xcf, 0xc8, 0x63, 0xa7, 0x73, 0x1b, 0xc3, 0xad, 0xe1, 0xf6,
	0xc3, 0x86, 0xdd, 0x62, 0xbc, 0xf4, 0x7a, 0x45, 0xca, 0xba, 0xb4, 0xd9, 0xf1, 0x36, 0x9d, 0x02,
	0xf3, 0x0e, 0x6f, 0xdb, 0x5f, 0x5b, 0xe5, 0x83, 0x5d, 0x37, 0xfd, 0x66, 0x4e, 0x7f, 0x05, 0x00,
	0x00, 0xff, 0xff, 0x34, 0x60, 0x58, 0x4d, 0xfa, 0x03, 0x00, 0x00,
}
