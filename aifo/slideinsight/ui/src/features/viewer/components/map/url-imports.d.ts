// Type declarations for Vite URL imports
declare module "*.js?url" {
  const url: string;
  export default url;
}

declare module "*.wasm?url" {
  const url: string;
  export default url;
}
