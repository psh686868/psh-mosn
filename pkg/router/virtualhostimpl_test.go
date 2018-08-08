package router

import (
	"strings"
	"testing"

	"github.com/psh686868/psh-mosn/pkg/api/v2"
	"github.com/psh686868/psh-mosn/pkg/protocol"
)

// Prefix > Path > Regex
func TestRouterPriority(t *testing.T) {
	// only prefix is valid
	prefixVitrualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{v2.Router{
			Match: v2.RouterMatch{
				Prefix: "/foo",
				Path:   "/foo.html",
				Regex:  ".*",
			},
			Route: v2.RouteAction{ClusterName: "test"},
		}},
	}, false)
	if len(prefixVitrualHost.routes) != 1 {
		t.Errorf("routes should have only one, but got :%v\n", len(prefixVitrualHost.routes))
	}
	if _, ok := prefixVitrualHost.routes[0].(*PrefixRouteRuleImpl); !ok {
		t.Error("cannot get a prefix route rule")
	}
	// only path is valid
	pathVirtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{v2.Router{
			Match: v2.RouterMatch{
				Path:  "/foo.html",
				Regex: ".*",
			},
			Route: v2.RouteAction{ClusterName: "test"},
		}},
	}, false)
	if len(pathVirtualHost.routes) != 1 {
		t.Errorf("routes should have only one, but got :%v\n", len(pathVirtualHost.routes))
	}
	if _, ok := pathVirtualHost.routes[0].(*PathRouteRuleImpl); !ok {
		t.Error("cannot get a path route rule")
	}
}

// the first matched route will be used
func TestRouterOrder(t *testing.T) {
	prefixrouter := v2.Router{
		Match: v2.RouterMatch{
			Prefix: "/foo",
		},
		Route: v2.RouteAction{ClusterName: "prefix"},
	}
	pathrouter := v2.Router{
		Match: v2.RouterMatch{
			Path: "/foo1",
		},
		Route: v2.RouteAction{ClusterName: "path"},
	}
	regrouter := v2.Router{
		Match: v2.RouterMatch{
			Regex: "/foo[0-9]+",
		},
		Route: v2.RouteAction{ClusterName: "regexp"},
	}
	// path "/foo1" match all of the router, the path router should be matched
	// path "/foo11" match prefix and regexp router, the regexp router should be matched
	// path "/foo" match prefix router only
	routers := []v2.Router{pathrouter, regrouter, prefixrouter}
	testCases := []struct {
		path        string
		clustername string
	}{
		{"/foo1", "path"},
		{"/foo11", "regexp"},
		{"/foo", "prefix"},
	}
	virtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: routers,
	}, false)
	for i, tc := range testCases {
		headers := map[string]string{
			strings.ToLower(protocol.MosnHeaderPathKey): tc.path,
		}
		rt := virtualHost.GetRouteFromEntries(headers, 1)
		if rt == nil || rt.RouteRule().ClusterName() != tc.clustername {
			t.Errorf("#%d route unexpected result\n", i)
		}
	}
	//prefix router first, only prefix will be matched
	prefixVirtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{prefixrouter, regrouter, pathrouter},
	}, false)
	for i, tc := range testCases {
		headers := map[string]string{
			strings.ToLower(protocol.MosnHeaderPathKey): tc.path,
		}
		rt := prefixVirtualHost.GetRouteFromEntries(headers, 1)
		if rt == nil || rt.RouteRule().ClusterName() != "prefix" {
			t.Errorf("#%d route unexpected result\n", i)
		}
	}

}
