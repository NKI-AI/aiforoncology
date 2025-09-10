// Lightweight runtime-toggled debug logging for the viewer

export function isDebugEnabled(): boolean {
  try {
    // Enable by running in DevTools: localStorage.setItem('slideDebug', '1')
    if (typeof window !== "undefined") {
      if ((window as any).__SLIDE_DEBUG__ === true) return true;
      const flag = window.localStorage.getItem("slideDebug");
      return flag === "1" || flag === "true" || flag === "on";
    }
  } catch (_) {
    // ignore
  }
  return false;
}

export function dbg(...args: any[]) {
  if (isDebugEnabled()) {
    // eslint-disable-next-line no-console
    console.log("🧪 DBG:", ...args);
  }
}
