/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"regexp"
	"strings"
	"time"

	"github.com/psh686868/psh-mosn/pkg/api/v2"
	"github.com/psh686868/psh-mosn/pkg/log"
	multimap "github.com/jwangsadinata/go-multimap/slicemultimap"
	//"github.com/psh686868/psh-mosn/pkg/protocol"
	"github.com/psh686868/psh-mosn/pkg/protocol"
	httpmosn "github.com/psh686868/psh-mosn/pkg/protocol/http"
	"github.com/psh686868/psh-mosn/pkg/types"
)

func NewRouteRuleImplBase(vHost *VirtualHostImpl, route *v2.Router) RouteRuleImplBase {
	routeRuleImplBase := RouteRuleImplBase{
		vHost:        vHost,
		routerMatch:  route.Match,
		routerAction: route.Route,
		policy: &routerPolicy{
			retryOn:      false,
			retryTimeout: 0,
			numRetries:   0,
		},
	}
	// todo add header match to route base
	// generate metadata match criteria from router's metadata
	if len(route.Route.MetadataMatch) > 0 {
		envoyLBMetaData := GetMosnLBMetaData(route)
		routeRuleImplBase.metadataMatchCriteria = NewMetadataMatchCriteriaImpl(envoyLBMetaData)

		routeRuleImplBase.metaData = GetClusterMosnLBMetaDataMap(route.Route.MetadataMatch)
	}

	return routeRuleImplBase
}

// Base implementation for all route entries.
type RouteRuleImplBase struct {
	caseSensitive               bool
	prefixRewrite               string
	hostRewrite                 string
	includeVirtualHostRateLimit bool
	corsPolicy                  types.CorsPolicy //todo
	vHost                       *VirtualHostImpl

	autoHostRewrite             bool
	useWebSocket                bool
	clusterName                 string //
	clusterHeaderName           lowerCaseString
	clusterNotFoundResponseCode httpmosn.Code
	timeout                     time.Duration
	runtime                     v2.RuntimeUInt32
	hostRedirect                string
	pathRedirect                string
	httpsRedirect               bool
	retryPolicy                 *retryPolicyImpl
	rateLimitPolicy             *rateLimitPolicyImpl

	routerAction v2.RouteAction
	routerMatch  v2.RouterMatch

	shadowPolicy          *shadowPolicyImpl
	priority              types.ResourcePriority
	configHeaders         []*types.HeaderData //
	configQueryParameters []types.QueryParameterMatcher
	weightedClusters      []*weightedClusterEntry
	totalClusterWeight    uint64
	hashPolicy            hashPolicyImpl

	metadataMatchCriteria *MetadataMatchCriteriaImpl
	metaData              types.RouteMetaData

	requestHeadersParser  *headerParser
	responseHeadersParser *headerParser

	opaqueConfig multimap.MultiMap

	decorator          *types.Decorator
	directResponseCode httpmosn.Code
	directResponseBody string
	policy             *routerPolicy
	virtualClusters    *VirtualClusterEntry
}

// types.RouterInfo
func (rri *RouteRuleImplBase) GetRouterName() string {

	return ""
}

// types.Route
func (rri *RouteRuleImplBase) RedirectRule() types.RedirectRule {

	return rri.RedirectRule()
}

func (rri *RouteRuleImplBase) RouteRule() types.RouteRule {

	return rri
}

func (rri *RouteRuleImplBase) TraceDecorator() types.TraceDecorator {

	return nil
}

// types.RouteRule
// Select Cluster for Routing
// todo support weighted cluster
func (rri *RouteRuleImplBase) ClusterName() string {

	return rri.routerAction.ClusterName
}

func (rri *RouteRuleImplBase) GlobalTimeout() time.Duration {

	return rri.routerAction.Timeout
}

func (rri *RouteRuleImplBase) Priority() types.Priority {

	return 0
}

func (rri *RouteRuleImplBase) VirtualHost() types.VirtualHost {

	return rri.vHost
}

func (rri *RouteRuleImplBase) VirtualCluster(headers map[string]string) types.VirtualCluster {

	return rri.virtualClusters
}

func (rri *RouteRuleImplBase) Policy() types.Policy {

	return rri.policy
}

func (rri *RouteRuleImplBase) Metadata() types.RouteMetaData {
	return rri.metaData
}

func (rri *RouteRuleImplBase) MetadataMatchCriteria() types.MetadataMatchCriteria {
	return rri.metadataMatchCriteria
}

// todo
func (rri *RouteRuleImplBase) finalizePathHeader(headers map[string]string, matchedPath string) {

}

func (rri *RouteRuleImplBase) matchRoute(headers map[string]string, randomValue uint64) bool {
	// todo check runtime
	// 1. match headers' KV
	if !ConfigUtilityInst.MatchHeaders(headers, rri.configHeaders) {
		return false
	}

	// 2. match query parameters
	var queryParams types.QueryParams

	if QueryString, ok := headers[protocol.MosnHeaderQueryStringKey]; ok {
		queryParams = httpmosn.ParseQueryString(QueryString)
	}

	if len(queryParams) == 0 {
		return true
	}

	return ConfigUtilityInst.MatchQueryParams(queryParams, rri.configQueryParameters)
}

type SofaRouteRuleImpl struct {
	RouteRuleImplBase
	matchValue string
}

func (srri *SofaRouteRuleImpl) Matcher() string {

	return srri.matchValue
}

func (srri *SofaRouteRuleImpl) MatchType() types.PathMatchType {

	return types.SofaHeader
}

func (srri *SofaRouteRuleImpl) Match(headers map[string]string, randomValue uint64) types.Route {
	if value, ok := headers[types.SofaRouteMatchKey]; ok {
		if value == srri.matchValue || srri.matchValue == ".*" {
			log.DefaultLogger.Debugf("Sofa router matches success")
			return srri
		}

		log.DefaultLogger.Warnf(" Sofa router matches failure, service name = %s", value)
	}

	log.DefaultLogger.Debugf("No service key found in header, sofa router matcher failure")

	return nil
}

type PathRouteRuleImpl struct {
	RouteRuleImplBase
	path string
}

func (prri *PathRouteRuleImpl) Matcher() string {

	return prri.path
}

func (prri *PathRouteRuleImpl) MatchType() types.PathMatchType {

	return types.Exact
}

// Exact Path Comparing
func (prri *PathRouteRuleImpl) Match(headers map[string]string, randomValue uint64) types.Route {
	// match base rule first
	log.StartLogger.Debugf("path route rule match invoked")
	if prri.matchRoute(headers, randomValue) {

		if headerPathValue, ok := headers[strings.ToLower(protocol.MosnHeaderPathKey)]; ok {

			if prri.caseSensitive {
				if headerPathValue == prri.path {
					log.DefaultLogger.Debugf("path route rule match success in caseSensitive scene")
					return prri
				}
			} else if strings.EqualFold(headerPathValue, prri.path) {
				log.DefaultLogger.Debugf("path route rule match success with exact matching ")
				return prri
			}
		}
	}
	log.DefaultLogger.Debugf("path route rule match failed")

	return nil
}

// todo
func (prri *PathRouteRuleImpl) FinalizeRequestHeaders(headers map[string]string, requestInfo types.RequestInfo) {
	prri.finalizePathHeader(headers, prri.path)
}

type PrefixRouteRuleImpl struct {
	RouteRuleImplBase
	prefix string
}

func (prei *PrefixRouteRuleImpl) Matcher() string {

	return prei.prefix
}

func (prei *PrefixRouteRuleImpl) MatchType() types.PathMatchType {

	return types.Prefix
}

// Compare Path's Prefix
func (prei *PrefixRouteRuleImpl) Match(headers map[string]string, randomValue uint64) types.Route {

	if prei.matchRoute(headers, randomValue) {

		if headerPathValue, ok := headers[strings.ToLower(protocol.MosnHeaderPathKey)]; ok {

			if strings.HasPrefix(headerPathValue, prei.prefix) {
				log.DefaultLogger.Debugf("prefix route rule match success")
				return prei
			}
		}
	}
	log.DefaultLogger.Debugf("prefix route rule match failed")

	return nil
}

func (prei *PrefixRouteRuleImpl) FinalizeRequestHeaders(headers map[string]string, requestInfo types.RequestInfo) {
	prei.finalizePathHeader(headers, prei.prefix)
}

//
type RegexRouteRuleImpl struct {
	RouteRuleImplBase
	regexStr     string
	regexPattern regexp.Regexp
}

func (rrei *RegexRouteRuleImpl) Matcher() string {

	return rrei.regexStr
}

func (rrei *RegexRouteRuleImpl) MatchType() types.PathMatchType {

	return types.Regex
}

func (rrei *RegexRouteRuleImpl) Match(headers map[string]string, randomValue uint64) types.Route {
	if rrei.matchRoute(headers, randomValue) {
		if headerPathValue, ok := headers[strings.ToLower(protocol.MosnHeaderPathKey)]; ok {

			if rrei.regexPattern.MatchString(headerPathValue) {
				log.DefaultLogger.Debugf("regex route rule match success")

				return rrei
			}
		}
	}
	log.DefaultLogger.Debugf("regex route rule match failed")

	return nil
}

func (rrei *RegexRouteRuleImpl) FinalizeRequestHeaders(headers map[string]string, requestInfo types.RequestInfo) {
	rrei.finalizePathHeader(headers, rrei.regexStr)
}
