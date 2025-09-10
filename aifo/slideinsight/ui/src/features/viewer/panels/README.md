# Panel Registry System

This directory contains the modular panel registry system for SlideInsight. The system allows for clean separation of concerns and easy addition of new panels to the workspace.

## Overview

The panel registry system provides:

1. **Modular Panel Architecture**: Each panel is a self-contained component with its own state management
2. **Automatic Dock Integration**: Panels automatically get dock buttons and keyboard shortcuts
3. **State Persistence**: Panel states (open/closed, dock position, size) are automatically persisted to localStorage
4. **Flexible Docking**: Panels can be docked to the left sidebar or float as overlays
5. **Context Sharing**: All panels have access to shared workspace context (map, slide data, etc.)

## Core Components

### Types (`types.ts`)

- `PanelRegistration`: Interface for registering new panels
- `PanelProps`: Props passed to each panel component
- `PanelContext`: Shared context available to all panels
- `PanelState`: Panel state management interface

### Registry (`PanelRegistry.tsx`)

- `PanelRegistryProvider`: Context provider for the panel system
- `usePanelRegistry`: Hook to access the full registry
- `usePanelManager`: Hook to access just the panel manager
- `usePanelContext`: Hook to access just the shared context

### Dock (`PanelDock.tsx`)

- `PanelDock`: Renders toggle buttons for all registered panels
- `PanelRenderer`: Renders all open panels (both docked and floating)
- `useDockedPanels`: Hook to get information about docked panels for layout

## Adding a New Panel

To add a new panel to the workspace:

### 1. Create Your Panel Component

```tsx
import React from "react";
import { PanelProps, PanelRegistration } from "@/features/viewer/panels/types";
import { YourIcon } from "@heroicons/react/24/outline";
import SlidePanel from "@/components/ui/slide-panel";

function YourPanelComponent({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  return (
    <SlidePanel
      isOpen={state.isOpen}
      onClose={onClose}
      dockOverride={state.dock}
      onDockChange={(dock) => updateState({ dock })}
      storageKey="yourPanel"
      defaultSize={state.size}
    >
      <SlidePanel.Header title="Your Panel" onClose={onClose} />

      <div className="flex-1 min-h-0 overflow-hidden">
        {/* Your panel content */}
        <div className="p-3">
          <p>Map ready: {context.mapRef ? "Yes" : "No"}</p>
          <p>Slide: {context.slideUid}</p>
          {/* Add your panel functionality here */}
        </div>
      </div>
    </SlidePanel>
  );
}

export const yourPanelRegistration: PanelRegistration = {
  id: "your-panel",
  name: "Your Panel",
  icon: YourIcon,
  component: YourPanelComponent,
  defaultState: {
    isOpen: false,
    dock: "left",
    size: { width: 320, height: 400 },
  },
  enabled: true,
  shortcut: "y", // Optional keyboard shortcut
  order: 50, // Display order in dock
};
```

### 2. Register Your Panel

```tsx
import { usePanelManager } from "@/features/viewer/panels";
import { yourPanelRegistration } from "./YourPanel";

function SomewhereInYourApp() {
  const manager = usePanelManager();

  useEffect(() => {
    manager.register(yourPanelRegistration);

    // Cleanup when component unmounts
    return () => {
      manager.unregister("your-panel");
    };
  }, [manager]);
}
```

### 3. Use Panel Context

Your panel has access to shared workspace context:

```tsx
function YourPanelComponent({
  context,
  state,
  updateState,
  onClose,
}: PanelProps) {
  const { mapRef, slideUid, studyUid, slideMetadata, rawSlideMetadata } =
    context;

  // Use the OpenLayers map
  useEffect(() => {
    if (mapRef) {
      // Add interactions, layers, etc.
      console.log("Map is ready!", mapRef);
    }
  }, [mapRef]);

  // Store custom state
  const handleSomething = () => {
    updateState({
      customState: {
        ...state.customState,
        yourCustomProperty: "some value",
      },
    });
  };
}
```

## Built-in Panels

The system includes wrappers for existing panels:

- **Mask Control Panel** (`mask-control`): Annotation management, shortcut: `A`
- **Annotation Editor** (`annotation-editor`): Annotation editing tools, shortcut: `E`
- **Brightness & Contrast** (`brightness-contrast`): Channel controls, shortcut: `B`

## Panel State Management

Each panel gets automatic state management:

```tsx
interface PanelState {
  isOpen: boolean; // Whether panel is open
  dock: "free" | "left"; // Docking mode
  size?: { width: number; height: number }; // Panel size (for floating)
  customState?: Record<string, any>; // Your custom state
}
```

State is automatically persisted to localStorage with the key `panel_{panelId}_state`.

## Keyboard Shortcuts

- Panels can define keyboard shortcuts in their registration
- `Escape` closes all panels
- Shortcuts only work when not focused on input/textarea elements

## Example Usage

See `examples/RegionPanelExample.tsx` for a complete example of creating a new panel.

## Integration with SlideWorkspace

The panel system is integrated into `SlideWorkspaceWithPanels.tsx`:

1. `PanelRegistryProvider` wraps the workspace
2. `PanelDock` renders in the left sidebar
3. `PanelRenderer` handles both docked and floating panels
4. Built-in panels are automatically registered on mount

## Migration from Old System

The old system had panels tightly coupled to SlideWorkspace. The new system:

- ✅ Separates panel logic from workspace logic
- ✅ Makes panels reusable and testable
- ✅ Provides automatic state management
- ✅ Enables easy addition of new panels
- ✅ Maintains backward compatibility through wrappers
