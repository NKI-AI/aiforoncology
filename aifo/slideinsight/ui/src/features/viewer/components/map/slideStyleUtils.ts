// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

export interface SlideStyleOptions {
  showGrayscale: boolean;
  invertBackground: boolean;
  gamma: number;
  channels?: Record<
    string,
    {
      visible: boolean;
      min: number;
      max: number;
    }
  >;
  // Legacy support for backward compatibility
  channelMin?: number;
  channelMax?: number;
  // Fluorescent metadata for proper channel blending
  channelMetadata?: Record<
    string,
    {
      name: string;
      biomarker: string;
      color: string; // Hex color like "#0000ff"
    }
  >;
}

/**
 * Generate WebGL style for fluorescent channel blending
 */
function generateFluorescentStyle(
  channelParams: Record<string, { visible: boolean; min: number; max: number }>,
  channelMetadata: Record<
    string,
    { name: string; biomarker: string; color: string }
  >,
  gamma: number,
  showGrayscale: boolean = false,
  invertBackground: boolean = false
): any {
  // Parse channel colors
  const parseHexColor = (hex: string): [number, number, number] => {
    const cleanHex = hex.replace("#", "");
    return [
      parseInt(cleanHex.substr(0, 2), 16) / 255.0,
      parseInt(cleanHex.substr(2, 2), 16) / 255.0,
      parseInt(cleanHex.substr(4, 2), 16) / 255.0,
    ];
  };

  // Collect contributions for each RGB component
  const redContributions: any[] = [];
  const greenContributions: any[] = [];
  const blueContributions: any[] = [];

  // Process each channel
  Object.keys(channelMetadata).forEach((channelId) => {
    const channelInfo = channelMetadata[channelId];
    const params = channelParams[channelId];

    // Skip invisible channels
    if (!params || !params.visible) {
      return;
    }

    // Get channel color
    const [r, g, b] = parseHexColor(channelInfo.color);

    // Create band index (convert string to number, add 1 for 1-based OpenLayers bands)
    const bandIndex = parseInt(channelId) + 1;

    // Create min/max normalization expression
    const normalizedMin = params.min / 65535.0;
    const normalizedMax = params.max / 65535.0;

    // First clamp the raw band value to [min, max] range, then rescale to [0,1]
    // Formula: (clamp(value, min, max) - min) / (max - min)
    const clampedValue = [
      "clamp",
      ["band", bandIndex],
      normalizedMin,
      normalizedMax,
    ];
    const normalizedIntensity =
      normalizedMax > normalizedMin
        ? [
            "/",
            ["-", clampedValue, normalizedMin],
            normalizedMax - normalizedMin,
          ]
        : 0; // Avoid division by zero if min == max

    // Add this channel's contribution to each RGB component
    if (r > 0) redContributions.push(["*", normalizedIntensity, r]);
    if (g > 0) greenContributions.push(["*", normalizedIntensity, g]);
    if (b > 0) blueContributions.push(["*", normalizedIntensity, b]);
  });

  // Build final RGB expressions by summing contributions
  let redExpression: any = 0;
  let greenExpression: any = 0;
  let blueExpression: any = 0;

  if (redContributions.length > 0) {
    redExpression =
      redContributions.length === 1
        ? redContributions[0]
        : ["+", ...redContributions];
  }
  if (greenContributions.length > 0) {
    greenExpression =
      greenContributions.length === 1
        ? greenContributions[0]
        : ["+", ...greenContributions];
  }
  if (blueContributions.length > 0) {
    blueExpression =
      blueContributions.length === 1
        ? blueContributions[0]
        : ["+", ...blueContributions];
  }

  // Clamp final RGB values to [0,1]
  let redFinal = ["clamp", redExpression, 0, 1];
  let greenFinal = ["clamp", greenExpression, 0, 1];
  let blueFinal = ["clamp", blueExpression, 0, 1];

  // Apply grayscale conversion if requested
  if (showGrayscale) {
    const grayValue = [
      "+",
      ["*", redFinal, 0.299],
      ["+", ["*", greenFinal, 0.587], ["*", blueFinal, 0.114]],
    ];
    redFinal = grayValue;
    greenFinal = grayValue;
    blueFinal = grayValue;
  }

  // Apply color inversion if requested
  if (invertBackground) {
    redFinal = ["-", 1, redFinal];
    greenFinal = ["-", 1, greenFinal];
    blueFinal = ["-", 1, blueFinal];
  }

  const finalColor = [
    "array",
    redFinal,
    greenFinal,
    blueFinal,
    1, // Alpha
  ];

  return {
    color: finalColor,
    gamma: gamma,
  };
}

/**
 * Generate WebGL style for slide layer based on controls
 */
export function generateSlideStyle(options: SlideStyleOptions): any {
  const {
    showGrayscale,
    invertBackground,
    gamma,
    channels,
    channelMin = 0,
    channelMax = 65535,
    channelMetadata,
  } = options;

  // If we have fluorescent metadata and channel parameters, use fluorescent blending
  if (channelMetadata && channels && Object.keys(channelMetadata).length > 0) {
    return generateFluorescentStyle(
      channels,
      channelMetadata,
      gamma,
      showGrayscale,
      invertBackground
    );
  }

  // Use per-channel settings if available, otherwise fall back to legacy global settings
  const usePerChannelSettings = channels && Object.keys(channels).length > 0;

  if (usePerChannelSettings) {
    // For per-channel settings, we'll need to handle this differently
    // For now, we'll use the first visible channel's settings as a fallback
    // TODO: This could be enhanced to generate more complex expressions
    const visibleChannels = Object.values(channels!).filter((ch) => ch.visible);
    if (visibleChannels.length === 0) {
      // No visible channels, return transparent
      return {
        color: ["array", 0, 0, 0, 0],
        gamma: gamma,
      };
    }

    // Use the first visible channel's settings for now
    // This is a simplified approach - more complex blending could be implemented
    const firstChannel = visibleChannels[0];
    const normalizeMin = firstChannel.min / 65535;
    const normalizeMax = firstChannel.max / 65535;

    return generateStyleWithMinMax(
      showGrayscale,
      invertBackground,
      gamma,
      normalizeMin,
      normalizeMax
    );
  } else {
    // Legacy mode - use global min/max
    const normalizeMin = channelMin / 65535;
    const normalizeMax = channelMax / 65535;

    return generateStyleWithMinMax(
      showGrayscale,
      invertBackground,
      gamma,
      normalizeMin,
      normalizeMax
    );
  }
}

/**
 * Helper function to generate style with normalized min/max values
 */
function generateStyleWithMinMax(
  showGrayscale: boolean,
  invertBackground: boolean,
  gamma: number,
  normalizeMin: number,
  normalizeMax: number
): any {
  // Base color expression - get RGB bands
  let colorExpression: any[] = [
    "array",
    ["band", 1], // R
    ["band", 2], // G
    ["band", 3], // B
    1, // Alpha
  ];

  // Apply brightness/contrast (min/max adjustment)
  if (normalizeMin !== 0 || normalizeMax !== 1) {
    colorExpression = [
      "array",
      [
        "*",
        ["-", ["band", 1], normalizeMin],
        1 / (normalizeMax - normalizeMin),
      ],
      [
        "*",
        ["-", ["band", 2], normalizeMin],
        1 / (normalizeMax - normalizeMin),
      ],
      [
        "*",
        ["-", ["band", 3], normalizeMin],
        1 / (normalizeMax - normalizeMin),
      ],
      1,
    ];
  }

  // Apply grayscale conversion (luminance formula: 0.299*R + 0.587*G + 0.114*B)
  if (showGrayscale) {
    const gray = [
      "+",
      ["*", ["band", 1], 0.299],
      ["+", ["*", ["band", 2], 0.587], ["*", ["band", 3], 0.114]],
    ];
    colorExpression = ["array", gray, gray, gray, 1];
  }

  // Apply color inversion
  if (invertBackground) {
    if (showGrayscale) {
      const gray = [
        "+",
        ["*", ["band", 1], 0.299],
        ["+", ["*", ["band", 2], 0.587], ["*", ["band", 3], 0.114]],
      ];
      const invertedGray = ["-", 1, gray];
      colorExpression = ["array", invertedGray, invertedGray, invertedGray, 1];
    } else {
      colorExpression = [
        "array",
        ["-", 1, ["band", 1]], // Invert R
        ["-", 1, ["band", 2]], // Invert G
        ["-", 1, ["band", 3]], // Invert B
        1,
      ];
    }
  }

  return {
    color: colorExpression,
    gamma: gamma,
  };
}
