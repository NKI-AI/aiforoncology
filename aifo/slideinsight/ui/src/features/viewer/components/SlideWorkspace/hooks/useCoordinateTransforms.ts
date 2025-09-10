import { useCallback } from "react";

/**
 * Hook that provides coordinate transformation helpers between map space and pixel space.
 *
 * The transformations are based on the microns-per-pixel (mpp) value of the slide.
 * - Map space is expressed in meters.
 * - Pixel space is expressed in pixels.
 * - Relationship: map = pixel * mpp * 1e-6, and pixel = map / (mpp * 1e-6)
 */
export function useCoordinateTransforms(slideMpp?: number) {
  const transformMapToPixelCoordinates = useCallback(
    (geoJsonData: any) => {
      if (!geoJsonData?.features || !slideMpp) {
        return geoJsonData;
      }

      const transformCoordinate = (coord: number[]): [number, number] => {
        if (coord.length < 2) return [0, 0];
        const pixelX = coord[0] / (slideMpp * 1e-6);
        const pixelY = -coord[1] / (slideMpp * 1e-6); // Note the negative for Y
        return [pixelX, pixelY];
      };

      const transformCoordinates = (coords: any): any => {
        if (
          Array.isArray(coords) &&
          coords.length >= 2 &&
          typeof coords[0] === "number"
        ) {
          return transformCoordinate(coords);
        } else if (Array.isArray(coords)) {
          return coords.map(transformCoordinates);
        }
        return coords;
      };

      return {
        ...geoJsonData,
        features: geoJsonData.features.map((feature: any) => ({
          ...feature,
          geometry: feature.geometry
            ? {
                ...feature.geometry,
                coordinates: transformCoordinates(feature.geometry.coordinates),
              }
            : feature.geometry,
        })),
      };
    },
    [slideMpp]
  );

  const transformPixelToMapCoordinates = useCallback(
    (geoJsonData: any) => {
      if (!geoJsonData?.features || !slideMpp) {
        return geoJsonData;
      }

      const transformCoordinate = (coord: number[]): [number, number] => {
        if (coord.length < 2) return [0, 0];
        const mapX = coord[0] * slideMpp * 1e-6;
        const mapY = -(coord[1] * slideMpp * 1e-6); // Note the negative for Y
        return [mapX, mapY];
      };

      const transformCoordinates = (coords: any): any => {
        if (
          Array.isArray(coords) &&
          coords.length >= 2 &&
          typeof coords[0] === "number"
        ) {
          return transformCoordinate(coords);
        } else if (Array.isArray(coords)) {
          return coords.map(transformCoordinates);
        }
        return coords;
      };

      return {
        ...geoJsonData,
        features: geoJsonData.features.map((feature: any) => ({
          ...feature,
          geometry: feature.geometry
            ? {
                ...feature.geometry,
                coordinates: transformCoordinates(feature.geometry.coordinates),
              }
            : feature.geometry,
        })),
      };
    },
    [slideMpp]
  );

  return {
    transformMapToPixelCoordinates,
    transformPixelToMapCoordinates,
  };
}
