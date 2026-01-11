let resolveReady: (() => void) | null = null;
let readyPromise: Promise<void> | null = null;
let isLoading = false;

export function startLoading() {
  isLoading = true;
  readyPromise = new Promise((resolve) => {
    resolveReady = resolve;
  });
}

export function finishLoading() {
  isLoading = false;
  if (resolveReady) {
    resolveReady();
    resolveReady = null;
    readyPromise = null;
  }
}

export function getReadyPromise(): Promise<void> | null {
  return readyPromise;
}

export function isPageLoading(): boolean {
  return isLoading;
}

export function resetLoader() {
  isLoading = false;
  resolveReady = null;
  readyPromise = null;
}
