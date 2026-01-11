import { useState, useRef, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { uploadFile } from '../../lib/api';
import { Icons } from '../Icons';

interface UploadItem {
  file: File;
  progress: number;
  speed: number;
  status: 'pending' | 'uploading' | 'done' | 'error' | 'cancelled';
  controller: AbortController;
  startTime?: number;
  loaded: number;
}

interface Props {
  open: boolean;
  onClose: () => void;
  serverId: string;
  path: string;
  onComplete: () => void;
  initialFiles?: FileList | null;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MiB`;
}

export default function UploadModal({ open, onClose, serverId, path, onComplete, initialFiles }: Props) {
  const [items, setItems] = useState<UploadItem[]>([]);
  const [visible, setVisible] = useState(false);
  const [animate, setAnimate] = useState(false);
  const [closing, setClosing] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadingRef = useRef(false);
  const initializedRef = useRef(false);

  const totalSize = items.reduce((a, i) => a + i.file.size, 0);
  const loadedSize = items.reduce((a, i) => a + i.loaded, 0);
  const completed = items.filter(i => i.status === 'done').length;
  const allDone = items.length > 0 && items.every(i => i.status === 'done' || i.status === 'error' || i.status === 'cancelled');

  const addFiles = (files: FileList | null) => {
    if (!files) return;
    const newItems: UploadItem[] = Array.from(files).map(file => ({
      file,
      progress: 0,
      speed: 0,
      status: 'pending',
      controller: new AbortController(),
      loaded: 0,
    }));
    setItems(prev => [...prev, ...newItems]);
  };

  useEffect(() => {
    if (open) {
      setClosing(false);
      setVisible(true);
      requestAnimationFrame(() => requestAnimationFrame(() => setAnimate(true)));
      if (initialFiles && !initializedRef.current) {
        initializedRef.current = true;
        addFiles(initialFiles);
      }
    }
  }, [open, initialFiles]);

  useEffect(() => {
    if (!open && visible && !closing) {
      handleClose();
    }
  }, [open, visible, closing]);

  const handleClose = useCallback(() => {
    if (closing) return;
    setClosing(true);
    setAnimate(false);
    if (!allDone) {
      items.forEach(i => i.controller.abort());
      setItems(prev => prev.map(i => ({ ...i, status: 'cancelled' })));
    }
    setTimeout(() => {
      setVisible(false);
      setClosing(false);
      setItems([]);
      initializedRef.current = false;
      onClose();
    }, 200);
  }, [closing, allDone, items, onClose]);

  const processQueue = async () => {
    if (uploadingRef.current) return;
    uploadingRef.current = true;

    const pending = items.find(i => i.status === 'pending');
    if (!pending) {
      uploadingRef.current = false;
      return;
    }

    setItems(prev => prev.map(i => i === pending ? { ...i, status: 'uploading', startTime: Date.now() } : i));

    await uploadFile(
      serverId,
      path,
      pending.file,
      (loaded, total) => {
        const elapsed = (Date.now() - (pending.startTime || Date.now())) / 1000;
        const speed = elapsed > 0 ? loaded / elapsed : 0;
        setItems(prev => prev.map(i =>
          i.file === pending.file ? { ...i, progress: (loaded / total) * 100, speed, loaded } : i
        ));
      },
      pending.controller.signal
    );

    setItems(prev => prev.map(i =>
      i.file === pending.file ? { ...i, status: i.status === 'uploading' ? 'done' : i.status, progress: 100, loaded: i.file.size } : i
    ));

    uploadingRef.current = false;
  };

  useEffect(() => {
    if (items.some(i => i.status === 'pending')) {
      processQueue();
    }
    if (items.length > 0 && items.every(i => i.status === 'done')) {
      onComplete();
    }
  }, [items]);

  const cancelItem = (item: UploadItem) => {
    item.controller.abort();
    setItems(prev => prev.map(i => i === item ? { ...i, status: 'cancelled' } : i));
  };

  const cancelAll = () => {
    items.forEach(i => i.controller.abort());
    setItems(prev => prev.map(i => ({ ...i, status: 'cancelled' })));
  };

  if (!visible) return null;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className={`absolute inset-0 bg-black/60 transition-opacity duration-200 ${animate && !closing ? 'opacity-100' : 'opacity-0'}`} onClick={allDone ? handleClose : undefined} />
      <div className={`relative w-[480px] max-w-[96vw] rounded-xl bg-neutral-900 shadow-2xl transition-all duration-200 ${animate && !closing ? 'opacity-100 scale-100' : 'opacity-0 scale-95'}`}>
        <div className="px-6 pt-6">
          <div className="text-lg font-semibold text-neutral-100">Upload files</div>
          <div className="text-sm text-neutral-400">Destination: {path}</div>
        </div>
        <div className="p-6 space-y-4">
          <div className="flex items-center justify-between gap-2">
            <div className="text-sm text-neutral-300">
              {completed}/{items.length} completed
              <span className="ml-2 text-neutral-500">{formatSize(loadedSize)} / {formatSize(totalSize)}</span>
            </div>
            <div className="flex items-center gap-2">
              <input ref={fileInputRef} type="file" multiple className="hidden" onChange={e => addFiles(e.target.files)} />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="rounded-lg font-semibold bg-neutral-800 text-neutral-100 hover:bg-neutral-700 text-xs px-3 py-1.5"
              >
                Add more
              </button>
              <button
                onClick={cancelAll}
                disabled={allDone}
                className="rounded-lg font-medium text-neutral-400 hover:bg-neutral-800 hover:text-neutral-100 disabled:opacity-60 text-xs px-3 py-1.5"
              >
                Cancel all
              </button>
            </div>
          </div>

          <div className="h-2 w-full rounded-full bg-neutral-800 overflow-hidden">
            <div className="h-full bg-sky-500 transition-all" style={{ width: `${totalSize ? (loadedSize / totalSize) * 100 : 0}%` }} />
          </div>

          {items.length > 0 && (
            <div className="max-h-80 overflow-auto rounded-md border border-neutral-800 divide-y divide-neutral-800">
              {items.map((item, idx) => (
                <div key={idx} className="p-3 bg-neutral-900">
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium text-neutral-100">{item.file.name}</div>
                      <div className="mt-0.5 text-xs text-neutral-500">
                        {formatSize(item.loaded)} / {formatSize(item.file.size)}
                        {item.status === 'uploading' && <><span className="mx-1">•</span>{formatSize(item.speed)}/s</>}
                        <span className="mx-1">•</span>
                        {item.status === 'pending' && 'Waiting'}
                        {item.status === 'uploading' && 'Uploading'}
                        {item.status === 'done' && 'Done'}
                        {item.status === 'error' && 'Error'}
                        {item.status === 'cancelled' && 'Cancelled'}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {item.status === 'uploading' && (
                        <span className="inline-block rounded-full border-2 border-current border-t-transparent animate-spin h-5 w-5 text-neutral-400" />
                      )}
                      {item.status === 'done' && (
                        <Icons.check className="h-5 w-5 text-emerald-500" />
                      )}
                      {(item.status === 'pending' || item.status === 'uploading') && (
                        <button
                          onClick={() => cancelItem(item)}
                          className="rounded-lg font-medium text-neutral-500 hover:text-neutral-100 text-xs px-2 py-1"
                        >
                          Cancel
                        </button>
                      )}
                    </div>
                  </div>
                  <div className="mt-2 h-1.5 w-full rounded-full bg-neutral-800 overflow-hidden">
                    <div
                      className={`h-full transition-all ${item.status === 'done' ? 'bg-emerald-500' : item.status === 'cancelled' || item.status === 'error' ? 'bg-red-500' : 'bg-sky-500'}`}
                      style={{ width: `${item.progress}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="flex justify-end gap-2">
            <button
              onClick={handleClose}
              disabled={items.length > 0 && !allDone}
              className="rounded-lg font-medium text-neutral-400 hover:text-neutral-100 disabled:opacity-60 text-xs px-3 py-1.5"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
}
