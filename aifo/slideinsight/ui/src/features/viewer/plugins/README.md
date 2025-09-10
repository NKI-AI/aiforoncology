# SlideInsight Plugin System

A clean, class-based plugin architecture for extending the SlideInsight viewer with modular functionality.

## Architecture Overview

The plugin system is built around class-based inheritance with well-defined interfaces and capabilities:

```
IPlugin (interface)
├── BasePlugin (abstract class)
    ├── MapInteractionPlugin (abstract class)
    ├── LayerPlugin (abstract class)
    └── MapLayerPlugin (abstract class)
```

## Core Components

### Base Classes

- **`BasePlugin`** - Abstract base for all plugins with core functionality
- **`MapInteractionPlugin`** - For plugins that interact with the map (drawing, selecting)
- **`LayerPlugin`** - For plugins that manage map layers
- **`MapLayerPlugin`** - Combines both map interactions and layer management

### Plugin Manager

- **`PluginManager`** - Handles plugin registration, lifecycle, and state management
- Supports capability-based plugin queries
- Provides proper initialization and cleanup
- Manages plugin state persistence

### Plugin Dock

- **`PluginDock`** - Renders plugin buttons in the viewer toolbar
- **`PluginRenderer`** - Renders plugin panels (docked and floating)
- **`useDockedPlugins`** - Hook for layout calculations

## Creating a Plugin

### 1. Simple Plugin

```typescript
import { BasePlugin, PluginAPI } from "./base";

class MyPlugin extends BasePlugin {
  public readonly id = "my-plugin";
  public readonly name = "My Plugin";
  public readonly version = "1.0.0";

  protected setupDefaultCapabilities(): void {
    this.addCapability("hasButton", true);
    this.addCapability("hasPanel", true);

    this.setButton({
      id: "my-plugin-button",
      label: "My Plugin",
      icon: MyIcon,
      tooltip: "Does something awesome",
      position: "right",
      order: 1,
    });

    this.setPanel({
      id: "my-plugin-panel",
      title: "My Plugin Panel",
      defaultSize: { width: 320, height: 400 },
      defaultDock: "free",
      storageKey: "myPluginPanel",
    });
  }

  protected createPanelComponent(): React.ComponentType<any> {
    return ({ api, onClose }) => (
      <div className="p-4">
        <h2>My Plugin Panel</h2>
        <p>Slide UID: {api.context.slideUid}</p>
        <button onClick={onClose}>Close</button>
      </div>
    );
  }
}
```

### 2. Map Interaction Plugin

```typescript
import { MapInteractionPlugin, MapInteractionContext } from "./base";

class MyMapPlugin extends MapInteractionPlugin {
  public readonly id = "my-map-plugin";
  public readonly name = "My Map Plugin";
  public readonly version = "1.0.0";

  protected setupDefaultCapabilities(): void {
    this.addCapability("hasMapInteractions", true);
    this.addCapability("hasButton", true);

    this.setButton({
      id: "my-map-button",
      label: "Map Tool",
      tooltip: "Interact with the map",
      position: "right",
      order: 2,
    });
  }

  public async setupMapInteractions(
    context: MapInteractionContext
  ): Promise<void> {
    const drawInteraction = new Draw({
      /* ... */
    });
    this.addInteraction(drawInteraction);
  }

  public async cleanupMapInteractions(): Promise<void> {
    // Cleanup handled automatically by base class
  }
}
```

### 3. Complex Plugin with Layers

```typescript
import { MapLayerPlugin, MapInteractionContext, LayerContext } from "./base";

class MyComplexPlugin extends MapLayerPlugin {
  public readonly id = "my-complex-plugin";
  public readonly name = "My Complex Plugin";
  public readonly version = "1.0.0";

  private myLayer?: VectorLayer<VectorSource>;

  protected setupDefaultCapabilities(): void {
    this.addCapability("hasMapInteractions", true);
    this.addCapability("hasLayers", true);
    this.addCapability("hasPanel", true);
  }

  public async createLayers(context: LayerContext): Promise<void> {
    const source = new VectorSource();
    this.myLayer = new VectorLayer({ source });
    this.addLayer(this.myLayer, 1500);
  }

  public async cleanupLayers(): Promise<void> {
    this.myLayer = undefined;
  }

  public async setupMapInteractions(
    context: MapInteractionContext
  ): Promise<void> {
    if (this.myLayer) {
      const drawInteraction = new Draw({
        source: this.myLayer.getSource()!,
        type: "Point",
      });
      this.addInteraction(drawInteraction);
    }
  }

  public async cleanupMapInteractions(): Promise<void> {
    // Cleanup handled by base class
  }
}
```

## Plugin Registration

```typescript
import { pluginManager } from "./base";

// Create and register plugin
const myPlugin = new MyPlugin();
await pluginManager.registerPlugin(
  myPlugin,
  {
    capabilities: {
      hasButton: true,
      hasPanel: true,
    },
  },
  {
    author: "Your Name",
    description: "Plugin description",
    category: "utility",
  }
);
```

## Plugin Capabilities

Plugins declare their capabilities to enable proper validation and querying:

```typescript
interface PluginCapabilities {
  hasMapInteractions?: boolean; // Can interact with map
  hasLayers?: boolean; // Manages map layers
  hasPanel?: boolean; // Has UI panel
  hasButton?: boolean; // Has toolbar button
  requiresSlideMetadata?: boolean; // Needs slide metadata
  requiresStudyContext?: boolean; // Needs study context
  canManageAnnotations?: boolean; // Works with annotations
  canManageRegions?: boolean; // Works with regions
  hasKeyboardShortcuts?: boolean; // Provides shortcuts
  persistsState?: boolean; // Saves state to localStorage
  canExportImport?: boolean; // Can export/import data
}
```

## Plugin Lifecycle

1. **Registration** - Plugin registered with manager
2. **Validation** - Capabilities and configuration validated
3. **Initialization** - `initialize()` called with plugin API
4. **Context Updates** - `onContextChange()` called when context changes
5. **User Interactions** - `onButtonClick()` called when button clicked
6. **Cleanup** - `destroy()` called when plugin unregistered

## Plugin Manager API

```typescript
// Register plugin
await pluginManager.registerPlugin(plugin, config, metadata);

// Query plugins
const allPlugins = pluginManager.getAllPlugins();
const mapPlugins = pluginManager.getPluginsByCapability("hasMapInteractions");
const annotationPlugins = pluginManager.getPluginsByCapability(
  "canManageAnnotations"
);

// Plugin state
const state = pluginManager.getPluginState(pluginId);
pluginManager.setPluginState(pluginId, { isOpen: true });

// Context updates
await pluginManager.updateContext({ slideUid: "new-slide" });

// Plugin lifecycle
await pluginManager.setPluginActive(pluginId, false);
await pluginManager.unregisterPlugin(pluginId);
```

## Examples

See the following files for complete examples:

- `BrightnessControlPluginV2.tsx` - Simple panel-based plugin
- `RegionControlPluginV2.tsx` - Complex map interaction and layer plugin
- `examples/PluginSystemDemo.tsx` - Complete demonstration

## Benefits

- **Clean Architecture**: Class-based inheritance with clear separation of concerns
- **Type Safety**: Full TypeScript support with comprehensive interfaces
- **Reusability**: Common functionality in base classes reduces duplication
- **Maintainability**: Clear structure makes plugins easier to understand and modify
- **Extensibility**: Easy to add new plugin types and capabilities
- **Validation**: Built-in validation and error handling
- **Performance**: Proper lifecycle management and cleanup
