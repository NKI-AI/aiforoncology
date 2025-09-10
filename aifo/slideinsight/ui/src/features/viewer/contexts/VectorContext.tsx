// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import {
  loadVectorSettings,
  saveVectorSettings,
  VectorContextSettings,
} from "@/features/viewer/utils/vectorLayerStorage";

interface VectorLayer {
  id: string;
  name: string;
  color: string;
  visible: boolean;
  defaultColor?: string; // Color extracted from GeoJSON features if available
  vectorName?: string; // Parent vector annotation name for grouping
}

interface VectorContextType {
  vectorOpacity: number;
  showVectors: boolean;
  vectorColors: string[]; // Default color palette
  vectorLayers: VectorLayer[]; // Individual vector layers with their info
  setVectorOpacity: (opacity: number) => void;
  setShowVectors: (show: boolean) => void;
  setVectorColors: (colors: string[]) => void;
  setVectorLayers: (layers: VectorLayer[]) => void;
  updateVectorLayer: (id: string, updates: Partial<VectorLayer>) => void;
  toggleVectorLayerVisibility: (id: string) => void;
}

const VectorContext = createContext<VectorContextType | undefined>(undefined);

export const useVectorContext = (): VectorContextType => {
  const context = useContext(VectorContext);
  if (!context) {
    throw new Error("useVectorContext must be used within a VectorProvider");
  }
  return context;
};

interface VectorProviderProps {
  children: ReactNode;
}

export const VectorProvider: React.FC<VectorProviderProps> = ({ children }) => {
  // Initialize state from localStorage
  const [vectorOpacity, setVectorOpacity] = useState<number>(() => {
    const settings = loadVectorSettings();
    return settings.vectorOpacity;
  });
  const [showVectors, setShowVectors] = useState<boolean>(() => {
    const settings = loadVectorSettings();
    return settings.showVectors;
  });
  const [vectorColors, setVectorColors] = useState<string[]>(() => {
    const settings = loadVectorSettings();
    return settings.vectorColors;
  });
  const [vectorLayers, setVectorLayers] = useState<VectorLayer[]>([]);

  const updateVectorLayer = (id: string, updates: Partial<VectorLayer>) => {
    setVectorLayers((prev) =>
      prev.map((layer) => (layer.id === id ? { ...layer, ...updates } : layer))
    );
  };

  const toggleVectorLayerVisibility = (id: string) => {
    setVectorLayers((prev) =>
      prev.map((layer) =>
        layer.id === id ? { ...layer, visible: !layer.visible } : layer
      )
    );
  };

  // Save settings to localStorage when they change
  useEffect(() => {
    saveVectorSettings({ vectorOpacity });
  }, [vectorOpacity]);

  useEffect(() => {
    saveVectorSettings({ showVectors });
  }, [showVectors]);

  useEffect(() => {
    saveVectorSettings({ vectorColors });
  }, [vectorColors]);

  return (
    <VectorContext.Provider
      value={{
        vectorOpacity,
        showVectors,
        vectorColors,
        vectorLayers,
        setVectorOpacity,
        setShowVectors,
        setVectorColors,
        setVectorLayers,
        updateVectorLayer,
        toggleVectorLayerVisibility,
      }}
    >
      {children}
    </VectorContext.Provider>
  );
};
