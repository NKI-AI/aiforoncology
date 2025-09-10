declare module "react-resizable-panels" {
  import * as React from "react";

  export interface PanelGroupProps
    extends React.HTMLAttributes<HTMLDivElement> {
    direction?: "horizontal" | "vertical";
    onLayout?: (sizes: number[]) => void;
  }

  export const PanelGroup: React.ForwardRefExoticComponent<
    PanelGroupProps & React.RefAttributes<HTMLDivElement>
  >;

  export interface PanelProps extends React.HTMLAttributes<HTMLDivElement> {
    defaultSize?: number;
    minSize?: number;
    maxSize?: number;
    onResize?: () => void;
  }

  export const Panel: React.FC<PanelProps>;

  export type PanelResizeHandleProps = React.HTMLAttributes<HTMLDivElement> & {
    withHandle?: boolean;
  };

  export const PanelResizeHandle: React.FC<PanelResizeHandleProps>;
}
