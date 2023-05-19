// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"fmt"
	"github.com/influxdata/influxdb-client-go/internal/http"
	"runtime"
)

const (
	Version = "1.4.0"
)

func init() {
	http.UserAgent = fmt.Sprintf("influxdb-client-go/%s  (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
}
