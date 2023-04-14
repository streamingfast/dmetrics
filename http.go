// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dmetrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func Serve(addr string) {
	serve := http.Server{Handler: promhttp.Handler(), Addr: addr}
	if err := serve.ListenAndServe(); err != nil {
		// It's common enough in development that we are good if it doesn't print
		zlog.Debug("can't listen on the metrics endpoint", zap.Error(err), zap.String("listen_addr", addr))
	}
}
