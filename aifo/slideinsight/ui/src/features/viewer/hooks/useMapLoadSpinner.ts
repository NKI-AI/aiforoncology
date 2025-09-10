// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import { useEffect, useState } from "react";
import type Map from "ol/Map";

export function useMapLoadSpinner(map: Map | null) {
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    if (!map) return;
    const onStart = () => setIsRefreshing(true);
    const onEnd = () => setIsRefreshing(false);
    map.on("loadstart", onStart);
    map.on("loadend", onEnd);
    return () => {
      map.un("loadstart", onStart);
      map.un("loadend", onEnd);
    };
  }, [map]);

  return isRefreshing;
}
