// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package datasources

import "aifo.dev/aifo/slideinsight/internal/server/ports"

// DataSources is a struct that contains all the data sources
// It is used to pass different data sources to the server and services
type DataSources struct {
	DB ports.Database
}
