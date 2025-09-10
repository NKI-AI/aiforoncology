import VectorSource from "ol/source/Vector";
import Feature from "ol/Feature";

export function getFeatureByIdSafe(
  source?: VectorSource,
  id?: string
): Feature | null {
  if (!source || !id) return null;
  try {
    const anySource = source as any;
    if (typeof anySource.getFeatureById === "function") {
      const f = anySource.getFeatureById(id) as Feature | null;
      if (f) return f;
    }
  } catch {}
  const all = source.getFeatures();
  return all.find((f) => String(f.getId?.()) === id) || null;
}

export function removeFeaturesById(
  source: VectorSource | undefined,
  ids: string[]
) {
  if (!source || !ids.length) return;
  ids.forEach((id) => {
    const f = getFeatureByIdSafe(source, id);
    if (f) {
      try {
        source.removeFeature(f);
      } catch {}
    }
  });
}

export function findInSourcesById(
  sources: Array<VectorSource | undefined>,
  id: string
): Feature | null {
  for (const s of sources) {
    const f = getFeatureByIdSafe(s, id);
    if (f) return f;
  }
  return null;
}

export function updateFeaturesById<T extends { id: string }>(
  source: VectorSource | undefined,
  items: T[],
  updateFn: (feature: Feature, item: T) => void
) {
  if (!source) return;
  items.forEach((item) => {
    const feat = getFeatureByIdSafe(source, item.id);
    if (feat) {
      updateFn(feat, item);
    }
  });
}
