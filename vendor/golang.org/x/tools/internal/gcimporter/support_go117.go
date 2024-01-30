// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !go1.18
// +build !go1.18

<<<<<<<< HEAD:vendor/golang.org/x/tools/internal/typeparams/enabled_go117.go
package typeparams

// Enabled reports whether type parameters are enabled in the current build
// environment.
const Enabled = false
========
package gcimporter

import "go/types"

const iexportVersion = iexportVersionGo1_11

func additionalPredeclared() []types.Type {
	return nil
}
>>>>>>>> daeedfd26 (Add latest version of `tektoncd` go client):vendor/golang.org/x/tools/internal/gcimporter/support_go117.go
