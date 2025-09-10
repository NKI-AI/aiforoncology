// Copyright 2025 Jonas Teuwen.
//
// Pluggable parsers for JXL payloads. Decouples decoding from tile source
// creation, so different decoders or WASM backends can be registered.

export type JxlDecoded = {
  data: Uint16Array | Uint8Array | Float32Array | Float64Array;
  width: number;
  height: number;
  channels: number;
  dataType: string;
  isMultiplex: boolean;
};

export type JxlParser = (
  buffer: ArrayBuffer,
  isMultiplex: boolean
) => Promise<JxlDecoded>;

const parserRegistry = new Map<string, JxlParser>();

export function registerJxlParser(name: string, parser: JxlParser): void {
  parserRegistry.set(name, parser);
}

export function unregisterJxlParser(name: string): void {
  parserRegistry.delete(name);
}

export function getJxlParser(name: string): JxlParser {
  const parser = parserRegistry.get(name);
  if (!parser) {
    throw new Error(`No JXL parser registered under name: ${name}`);
  }
  return parser;
}

export function hasJxlParser(name: string): boolean {
  return parserRegistry.has(name);
}
