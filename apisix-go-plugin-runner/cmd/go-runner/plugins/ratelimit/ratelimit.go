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

package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rhttp "github.com/Conflux-Chain/go-conflux-util/rate/http"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
)

var (
	// registries *rhttp.Registry
	serverReqRegistry  = rhttp.NewRegistry(&RainbowLimiterFactory{qpsObtainer: &serverQpsObtainer{}})
	serverCostRegistry = rhttp.NewRegistry(&RainbowLimiterFactory{qpsObtainer: &serverQpsObtainer{}})
	costRegistry       = rhttp.NewRegistry(&RainbowLimiterFactory{qpsObtainer: &costtypeQpsObtainer{}})
)

func init() {
	err := plugin.RegisterPlugin(&RateLimit{})
	if err != nil {
		log.Fatalf("failed to register plugin RateLimit: %s", err)
	}
}

// RateLimit is a demo to show how to return data directly instead of proxying
// it to the upstream.
type RateLimit struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type RateLimitConf struct {
	Mode string `json:"mode"` // request or cost_type
}

// func (r *RateLimitConf) GetRegistry() (*rhttp.Registry, error) {
// if r.Mode != "request" || r.Mode != "cost_type" {
// 	return nil, fmt.Errorf("not support mode %s", r.Mode)
// }

// if _, ok := registries[*r]; !ok {
// 	switch r.Mode {
// 	case "request":
// 		registries[*r] = rhttp.NewRegistry(&RequestLimiterFactory{})
// 	case "cost_type":
// 		return nil, errors.New("not implemented")
// 	default:
// 		return nil, fmt.Errorf("not support mode %s", r.Mode)
// 	}
// }
// return registries[*r], nil
// }

func (p *RateLimit) Name() string {
	return "rate-limit"
}

func (p *RateLimit) ParseConf(in []byte) (interface{}, error) {
	conf := RateLimitConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func (p *RateLimit) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	fn := func() error {
		serverType := r.Header().Get(constants.RAINBOW_SERVER_TYPE_HEADER_KEY)
		costType := r.Header().Get(constants.RAINBOW_COST_TYPE_HEADER_KEY)
		userId := r.Header().Get(constants.RAINBOW_USER_ID_HEADER_KEY)

		c := conf.(RateLimitConf)
		ctx := context.WithValue(context.Background(), constants.RAINBOW_USER_ID_HEADER_KEY, userId)
		switch c.Mode {
		case "request":
			return serverReqRegistry.Limit(ctx, serverType)
		case "cost_type":
			if err := serverCostRegistry.Limit(ctx, serverType); err != nil {
				return err
			}
			return costRegistry.Limit(ctx, costType)
		default:
			return fmt.Errorf("unsupport limit mode %s", c.Mode)
		}
	}

	if err := fn(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf("failed count for request: %v", err))); err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
