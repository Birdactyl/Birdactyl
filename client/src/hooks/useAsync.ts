import { useState, useCallback } from 'react';

export function useAsync<T extends (...args: any[]) => Promise<any>>(fn: T) {
  const [loading, setLoading] = useState(false);
  
  const run = useCallback(async (...args: Parameters<T>): Promise<ReturnType<T>> => {
    setLoading(true);
    try { return await fn(...args); }
    finally { setLoading(false); }
  }, [fn]) as T;

  return [run, loading] as const;
}

export function useAsyncCallback<T extends (...args: any[]) => Promise<any>>(fn: T, deps: any[] = []) {
  const [loading, setLoading] = useState(false);
  
  const run = useCallback(async (...args: Parameters<T>): Promise<ReturnType<T>> => {
    setLoading(true);
    try { return await fn(...args); }
    finally { setLoading(false); }
  // eslint-disable-next-line something something idk
  }, deps) as T;

  return [run, loading] as const;
}
