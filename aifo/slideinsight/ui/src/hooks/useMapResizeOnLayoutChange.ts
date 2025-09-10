import { useEffect } from "react";
import type Map from "ol/Map";

export function useMapResizeOnLayoutChange(
  map: Map | null,
  deps: any[],
  delay = 100
) {
  useEffect(() => {
    if (!map) return;
    const id = window.setTimeout(() => {
      try {
        map.updateSize();
      } catch {}
    }, delay);
    return () => window.clearTimeout(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [map, delay, ...deps]);
}
