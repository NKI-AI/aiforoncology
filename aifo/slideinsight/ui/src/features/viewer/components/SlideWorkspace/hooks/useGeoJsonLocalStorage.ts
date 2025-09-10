import { useMemo } from "react";

type Feature = any; // Using any to match existing flexible usage
type FeatureCollection = { type: "FeatureCollection"; features: Feature[] };

function safeParseFeatureCollection(raw: string | null): FeatureCollection {
  if (!raw) return { type: "FeatureCollection", features: [] };
  try {
    const data = JSON.parse(raw);
    if (!data || !Array.isArray(data.features)) {
      return { type: "FeatureCollection", features: [] };
    }
    return data;
  } catch {
    return { type: "FeatureCollection", features: [] };
  }
}

function saveFeatureCollection(key: string, fc: FeatureCollection): void {
  try {
    localStorage.setItem(key, JSON.stringify(fc));
  } catch {
    // ignore
  }
}

export function useGeoJsonLocalStorage(
  slideUid: string,
  transformMapToPixelCoordinates: (fc: FeatureCollection) => FeatureCollection,
  transformPixelToMapCoordinates: (fc: FeatureCollection) => FeatureCollection
) {
  const annotationsKey = useMemo(
    () => `slideAnnotations:${slideUid}`,
    [slideUid]
  );
  const regionsKey = useMemo(() => `slideRegions:${slideUid}`, [slideUid]);

  const loadCollection = (
    key: string,
    toMapSpace = false
  ): FeatureCollection => {
    const raw = localStorage.getItem(key);
    const fc = safeParseFeatureCollection(raw);
    return toMapSpace ? transformPixelToMapCoordinates(fc) : fc;
  };

  const appendFeatureInMapCoords = (
    key: string,
    featureInMapCoords: Feature
  ): void => {
    const existing = safeParseFeatureCollection(localStorage.getItem(key));
    const fcToSave = transformMapToPixelCoordinates({
      type: "FeatureCollection",
      features: [featureInMapCoords],
    });
    existing.features.push(fcToSave.features[0]);
    saveFeatureCollection(key, existing);
  };

  const upsertFeatureInMapCoordsById = (
    key: string,
    featureInMapCoords: Feature
  ): void => {
    const existing = safeParseFeatureCollection(localStorage.getItem(key));
    const fcToSave = transformMapToPixelCoordinates({
      type: "FeatureCollection",
      features: [featureInMapCoords],
    });
    const updated = fcToSave.features[0];
    const fid = String(updated?.id ?? "");
    if (!fid) return;

    const idx = existing.features.findIndex(
      (f: any) => f?.id != null && String(f.id) === fid
    );
    if (idx >= 0) {
      updated.properties = {
        ...(existing.features[idx].properties || {}),
        ...(updated.properties || {}),
      };
      existing.features[idx] = updated;
    }
    saveFeatureCollection(key, existing);
  };

  const removeByIds = (key: string, ids: string[]): void => {
    if (!ids.length) return;
    const existing = safeParseFeatureCollection(localStorage.getItem(key));
    existing.features = existing.features.filter((f: any) => {
      const fid = f?.id != null ? String(f.id) : undefined;
      return fid ? !ids.includes(fid) : true;
    });
    saveFeatureCollection(key, existing);
  };

  const updateProperties = (
    key: string,
    idToProps: Record<string, any>
  ): void => {
    const existing = safeParseFeatureCollection(localStorage.getItem(key));
    existing.features.forEach((f: any) => {
      const fid = f?.id != null ? String(f.id) : undefined;
      if (fid && idToProps[fid]) {
        if (!f.properties) f.properties = {};
        f.properties = { ...f.properties, ...idToProps[fid] };
      }
    });
    saveFeatureCollection(key, existing);
  };

  return {
    annotationsKey,
    regionsKey,
    loadCollection,
    appendFeatureInMapCoords,
    upsertFeatureInMapCoordsById,
    removeByIds,
    updateProperties,
  };
}
