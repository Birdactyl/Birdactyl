import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  listFiles, FileEntry, searchFiles, SearchResult, deleteFile, moveFile, copyFile,
  compressFile, decompressFile, createFolder, writeFile, bulkDeleteFiles, bulkCopyFiles, bulkCompressFiles
} from '../lib/api';
import { notify } from '../components/feedback/Notification';

export function useFileManager(serverId: string | undefined, initialPath: string) {
  const navigate = useNavigate();
  const [files, setFiles] = useState<FileEntry[]>([]);
  const [currentPath, setCurrentPath] = useState(initialPath);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [clipboard, setClipboard] = useState<string[]>([]);
  const [pasting, setPasting] = useState(false);
  const [decompressing, setDecompressing] = useState(false);
  const uploadInputRef = useRef<HTMLInputElement>(null);

  const getFilePath = (name: string) => currentPath === '/' ? `/${name}` : `${currentPath}/${name}`;

  const refreshFiles = useCallback(() => {
    if (!serverId) return;
    listFiles(serverId, currentPath).then(res => {
      if (res.success && res.data) {
        setFiles([...res.data].sort((a, b) => a.is_dir !== b.is_dir ? (a.is_dir ? -1 : 1) : a.name.localeCompare(b.name)));
        setError(null);
      } else if (res.error === 'Permission denied') {
        setError("You don't have permission to view files");
      }
    });
  }, [serverId, currentPath]);

  useEffect(() => {
    if (!serverId) return;
    setLoading(true);
    setSelected(new Set());
    listFiles(serverId, currentPath).then(res => {
      if (res.success && res.data) {
        setFiles([...res.data].sort((a, b) => a.is_dir !== b.is_dir ? (a.is_dir ? -1 : 1) : a.name.localeCompare(b.name)));
        setError(null);
      } else if (res.error === 'Permission denied') {
        setError("You don't have permission to view files");
      }
      setLoading(false);
    });
  }, [serverId, currentPath]);

  useEffect(() => {
    if (!serverId || !search.trim()) { setSearchResults(null); return; }
    const timer = setTimeout(() => {
      searchFiles(serverId, search).then(res => res.success && res.data && setSearchResults(res.data));
    }, 300);
    return () => clearTimeout(timer);
  }, [serverId, search]);

  const goUp = () => {
    if (currentPath === '/') return;
    const parts = currentPath.split('/').filter(Boolean);
    parts.pop();
    setCurrentPath(parts.length ? '/' + parts.join('/') : '/');
  };

  const navigateTo = (entry: FileEntry) => {
    if (entry.is_dir) {
      entry.name === '..' ? goUp() : setCurrentPath(getFilePath(entry.name));
      setSearch('');
    } else {
      navigate(`/console/server/${serverId}/files/edit?path=${encodeURIComponent(getFilePath(entry.name))}`);
    }
  };

  const navigateToSearchResult = (result: SearchResult) => {
    if (result.is_dir) { setCurrentPath(result.path); setSearch(''); }
    else navigate(`/console/server/${serverId}/files/edit?path=${encodeURIComponent(result.path)}`);
  };

  const checkPerm = (res: { success: boolean; error?: string }) => {
    if (res.error === 'Permission denied') {
      notify('Permission Denied', "You don't have permission to perform this action", 'error');
      return false;
    }
    return res.success;
  };

  const actions = {
    delete: async (file: FileEntry) => { if (serverId) { const res = await deleteFile(serverId, getFilePath(file.name)); checkPerm(res); refreshFiles(); } },
    move: async (file: FileEntry, dest: string) => { if (serverId) { const res = await moveFile(serverId, getFilePath(file.name), dest); checkPerm(res); refreshFiles(); } },
    rename: async (file: FileEntry, newName: string) => { if (serverId) { const res = await moveFile(serverId, getFilePath(file.name), getFilePath(newName)); checkPerm(res); refreshFiles(); } },
    copy: (file: FileEntry) => { const p = getFilePath(file.name); if (!clipboard.includes(p)) setClipboard([...clipboard, p]); },
    duplicate: async (file: FileEntry) => { if (serverId) { const res = await copyFile(serverId, getFilePath(file.name), getFilePath(`copy_${file.name}`)); checkPerm(res); refreshFiles(); } },
    compress: async (file: FileEntry, format: string) => {
      if (!serverId) return;
      const base = file.name.replace(/\.[^/.]+$/, '');
      const res = await compressFile(serverId, getFilePath(file.name), getFilePath(`${base}.${format === 'tar.gz' ? 'tar.gz' : format}`), format);
      checkPerm(res);
      refreshFiles();
    },
    decompress: async (file: FileEntry) => {
      if (!serverId || decompressing) return;
      setDecompressing(true);
      const res = await decompressFile(serverId, getFilePath(file.name), currentPath);
      checkPerm(res);
      setDecompressing(false);
      refreshFiles();
    },
    createFolder: async (name: string) => { if (serverId) { const res = await createFolder(serverId, getFilePath(name)); checkPerm(res); refreshFiles(); } },
    createFile: async (name: string) => {
      if (!serverId) return;
      const path = getFilePath(name);
      const res = await writeFile(serverId, path, '');
      if (res.error === 'Permission denied') { setError("You don't have permission to create files"); return; }
      navigate(`/console/server/${serverId}/files/edit?path=${encodeURIComponent(path)}`);
    },
    paste: async () => {
      if (!serverId || clipboard.length === 0 || pasting) return;
      setPasting(true);
      const res = await bulkCopyFiles(serverId, clipboard, currentPath);
      checkPerm(res);
      setClipboard([]);
      setPasting(false);
      refreshFiles();
    },
    bulkDelete: async () => {
      if (!serverId) return;
      const res = await bulkDeleteFiles(serverId, Array.from(selected).map(getFilePath));
      checkPerm(res);
      setSelected(new Set());
      refreshFiles();
    },
    bulkCopy: () => {
      const paths = Array.from(selected).map(getFilePath);
      setClipboard([...clipboard, ...paths.filter(p => !clipboard.includes(p))]);
      setSelected(new Set());
    },
    bulkCompress: async (format: string) => {
      if (!serverId) return;
      const paths = Array.from(selected).map(getFilePath);
      const res = await bulkCompressFiles(serverId, paths, getFilePath(`archive.${format === 'tar.gz' ? 'tar.gz' : format}`), format);
      checkPerm(res);
      setSelected(new Set());
      refreshFiles();
    },
  };

  const toggleSelect = (name: string) => {
    const next = new Set(selected);
    next.has(name) ? next.delete(name) : next.add(name);
    setSelected(next);
  };

  const selectableFiles = files.filter(f => f.name !== '..');
  const allSelected = selectableFiles.length > 0 && selected.size === selectableFiles.length;
  const someSelected = selected.size > 0 && selected.size < selectableFiles.length;
  const toggleAll = () => setSelected(allSelected ? new Set() : new Set(selectableFiles.map(f => f.name)));

  const parentEntry: FileEntry = { name: '..', size: 0, is_dir: true, mod_time: 0, mode: '' };
  const filteredFiles = [...(currentPath !== '/' ? [parentEntry] : []), ...files.filter(f => f.name.toLowerCase().includes(search.toLowerCase()))];

  return {
    files: filteredFiles, loading, error, currentPath, setCurrentPath, search, setSearch, searchResults,
    selected, toggleSelect, allSelected, someSelected, toggleAll, clipboard, setClipboard, pasting, decompressing,
    uploadInputRef, goUp, navigateTo, navigateToSearchResult, refreshFiles, getFilePath, actions,
  };
}
