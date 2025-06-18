// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package datasources

import "aifo.dev/aifo/slidescope/internal/datasources/database"

// DataSources is a struct that contains all the data sources
// It is used to pass different data sources to the server and services
type DataSources struct {
	DB database.Database
}
