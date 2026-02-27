import { useState, useEffect } from 'react';

import { useParams, useSearchParams } from 'react-router-dom';
import { getServer, Server, FileEntry, SearchResult, getDownloadUrl } from '../../../lib/api';
import { formatBytes, formatDate } from '../../../lib/utils';
import { useFileManager } from '../../../hooks/useFileManager';
import { useServerPermissions } from '../../../hooks/useServerPermissions';
import { UploadModal, CreateFolderModal, CreateFileModal, MoveFileModal, RenameFileModal, CompressFileModal, ClipboardPanel, Button, Icons, Checkbox, PermissionDenied, Input, DeleteFileModal, ContextMenuZone, BulkActionBar } from '../../../components';

const getFileIconColor = (name: string, isDir: boolean): string => {
  if (isDir) return 'text-amber-500';
  if (['tar', 'gz', 'zip', 'rar', '7z', 'bz2'].some(e => name.toLowerCase().includes(e))) return 'text-orange-500';
  if (['json', 'yaml', 'yml', 'xml', 'toml'].includes(name.split('.').pop()?.toLowerCase() || '')) return 'text-emerald-600';
  return 'text-neutral-500';
};

const FileIcon = ({ name, is_dir }: { name: string; is_dir: boolean }) => {
  const color = getFileIconColor(name, is_dir);
  if (is_dir) return <Icons.folder className={`w-5 h-5 ${color}`} />;
  if (['tar', 'gz', 'zip', 'rar', '7z', 'bz2'].some(e => name.toLowerCase().includes(e))) return <Icons.archive className={`w-5 h-5 ${color}`} />;
  if (['json', 'yaml', 'yml', 'xml', 'toml'].includes(name.split('.').pop()?.toLowerCase() || '')) return <Icons.fileText className={`w-5 h-5 ${color}`} />;
  return <Icons.file className={`w-5 h-5 ${color}`} />;
};

const isArchive = (name: string) => /\.(zip|tar|tar\.gz|tgz)$/i.test(name);

export default function FilesPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const [server, setServer] = useState<Server | null>(null);
  const fm = useFileManager(id, searchParams.get('path') || '/');
  const { can, loading: permsLoading } = useServerPermissions(id);

  const [modals, setModals] = useState<{ newFolder: boolean; newFile: boolean; upload: boolean; bulkCompress: boolean; initialFiles: FileList | null }>({ newFolder: false, newFile: false, upload: false, bulkCompress: false, initialFiles: null });
  const [fileTarget, setFileTarget] = useState<{ type: 'move' | 'rename' | 'compress'; file: FileEntry } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ file: FileEntry } | { bulk: true } | null>(null);

  useEffect(() => { id && getServer(id).then(res => res.success && res.data && setServer(res.data)); }, [id]);

  if (permsLoading) return null;
  if (!can('file.list')) return <PermissionDenied message="You don't have permission to view files" />;

  const handleDownload = (file: FileEntry) => {
    if (!id) return;
    const a = document.createElement('a');
    a.href = getDownloadUrl(id, fm.getFilePath(file.name));
    a.download = file.name;
    a.click();
  };

  const getFileActions = (file: FileEntry) => {
    if (file.name === '..') return [];
    if (file.is_dir) {
      return [
        ...(can('file.copy') ? [{ label: 'Move', onClick: () => setFileTarget({ type: 'move', file }) }] : []),
        ...(can('file.rename') ? [{ label: 'Rename', onClick: () => setFileTarget({ type: 'rename', file }) }] : []),
        ...(can('file.compress') ? [{ label: 'Compress', onClick: () => setFileTarget({ type: 'compress', file }) }] : []),
        'separator' as const,
        ...(can('file.delete') ? [{ label: 'Delete', onClick: () => setDeleteTarget({ file }), variant: 'danger' as const }] : []),
      ];
    }
    return [
      ...(can('file.read') ? [{ label: 'Edit', onClick: () => fm.navigateTo(file) }] : []),
      ...(can('file.read') ? [{ label: 'Download', onClick: () => handleDownload(file) }] : []),
      'separator' as const,
      ...(can('file.copy') ? [{ label: 'Copy', onClick: () => fm.actions.copy(file) }] : []),
      ...(can('file.copy') ? [{ label: 'Duplicate', onClick: () => fm.actions.duplicate(file) }] : []),
      ...(can('file.copy') ? [{ label: 'Move', onClick: () => setFileTarget({ type: 'move', file }) }] : []),
      ...(can('file.rename') ? [{ label: 'Rename', onClick: () => setFileTarget({ type: 'rename', file }) }] : []),
      'separator' as const,
      ...(can('file.compress') ? [{ label: 'Compress', onClick: () => setFileTarget({ type: 'compress', file }) }] : []),
      ...(isArchive(file.name) && can('file.compress') ? [{ label: fm.decompressing ? 'Extracting...' : 'Extract', onClick: () => fm.actions.decompress(file), disabled: fm.decompressing }] : []),
      'separator' as const,
      ...(can('file.delete') ? [{ label: 'Delete', onClick: () => setDeleteTarget({ file }), variant: 'danger' as const }] : []),
    ];
  };

  const getSearchActions = (result: SearchResult) => [
    { label: 'Open Directory', onClick: () => { const dir = result.path.substring(0, result.path.lastIndexOf('/')) || '/'; fm.setCurrentPath(dir); fm.setSearch(''); } },
  ];

  return (
    <div className="space-y-4">
      {fm.error && <PermissionDenied message={fm.error} />}

      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-neutral-100">Files</h1>
          <div className="flex items-center gap-2">
            <Button variant="ghost" onClick={fm.goUp} disabled={fm.currentPath === '/'}><Icons.arrowUp className="h-4 w-4" /></Button>
            <Button variant="ghost" onClick={fm.refreshFiles}><Icons.refresh className="h-4 w-4" /></Button>
          </div>
        </div>

        <div className="flex items-center gap-1 text-xs text-neutral-400 overflow-x-auto pb-1">
          <button type="button" className="hover:text-neutral-200 shrink-0" onClick={() => fm.setCurrentPath('/')}>{server?.name || 'Server'}</button>
          <span>/</span>
          <button type="button" className="hover:text-neutral-200 shrink-0" onClick={() => fm.setCurrentPath('/')}>files</button>
          {fm.currentPath !== '/' && fm.currentPath.split('/').filter(Boolean).map((part, i, arr) => (
            <span key={i} className="flex items-center gap-1 shrink-0">
              <span>/</span>
              <button type="button" className="hover:text-neutral-200" onClick={() => fm.setCurrentPath('/' + arr.slice(0, i + 1).join('/'))}>{part}</button>
            </span>
          ))}
        </div>

        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <Input className="w-full sm:flex-1" placeholder="Search files..." value={fm.search} onChange={e => fm.setSearch(e.target.value)} />
          <div className="flex items-center gap-3 overflow-x-auto scrollbar-hide px-1 py-1">
            {can('file.create') && <Button variant="ghost" onClick={() => setModals(m => ({ ...m, newFolder: true }))} className="shrink-0"><Icons.folderPlus className="h-4 w-4" /></Button>}
            {can('file.create') && <Button variant="ghost" onClick={() => setModals(m => ({ ...m, newFile: true }))} className="shrink-0"><Icons.filePlus className="h-4 w-4" /></Button>}
            <input ref={fm.uploadInputRef} type="file" multiple className="hidden" onChange={e => { if (e.target.files?.length) setModals(m => ({ ...m, upload: true, initialFiles: e.target.files })); }} />
            {can('file.upload') && <Button onClick={() => fm.uploadInputRef.current?.click()} className="shrink-0"><Icons.arrowUp className="h-4 w-4 sm:mr-1.5" /><span className="hidden sm:inline">Upload</span></Button>}
            {can('file.copy') && fm.clipboard.length > 0 && <Button onClick={fm.actions.paste} disabled={fm.pasting} loading={fm.pasting} className="shrink-0"><Icons.clipboardCheck className="h-4 w-4 sm:mr-1.5" /><span className="hidden sm:inline">{fm.pasting ? 'Pasting...' : 'Paste'}</span></Button>}
          </div>
        </div>
      </div>

      <div className="bg-neutral-900/40 rounded-lg border border-neutral-800 overflow-hidden">
        <div className="hidden md:block">
          <table className="min-w-full">
            <thead className="bg-neutral-900/50">
              <tr>
                <th className="w-10 pl-4 py-3"><Checkbox checked={fm.allSelected} indeterminate={fm.someSelected} onChange={fm.toggleAll} /></th>
                <th className="px-3 py-3 text-left text-xs font-medium text-neutral-500 uppercase">Name</th>
                <th className="px-3 py-3 text-left text-xs font-medium text-neutral-500 uppercase w-24">Size</th>
                <th className="px-3 py-3 text-left text-xs font-medium text-neutral-500 uppercase w-36">Modified</th>
                <th className="w-12 px-3 py-3"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800">
              {fm.loading ? (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-sm text-neutral-500">Loading...</td></tr>
              ) : fm.searchResults ? (
                fm.searchResults.length === 0 ? (
                  <tr><td colSpan={5} className="px-4 py-8 text-center text-sm text-neutral-500">No results found</td></tr>
                ) : fm.searchResults.map(result => (
                  <ContextMenuZone as="tr" key={result.path} items={getSearchActions(result)} className="hover:bg-neutral-800/50 cursor-pointer" onClick={() => fm.navigateToSearchResult(result)}>
                    <td className="pl-4 py-3"></td>
                    <td className="px-3 py-3">
                      <div className="flex items-center gap-3">
                        <FileIcon name={result.name} is_dir={result.is_dir} />
                        <div className="min-w-0">
                          <div className="text-sm text-neutral-100 truncate">{result.name}</div>
                          <div className="text-xs text-neutral-500 truncate">{result.path}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-3 py-3 text-sm text-neutral-400">{result.is_dir ? '\u2014' : formatBytes(result.size)}</td>
                    <td className="px-3 py-3 text-sm text-neutral-400">{formatDate(result.mod_time)}</td>
                    <td className="px-3 py-3"><Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button></td>
                  </ContextMenuZone>
                ))
              ) : fm.files.length === 0 ? (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-sm text-neutral-500">No files found</td></tr>
              ) : fm.files.map(file => {
                const actions = getFileActions(file);
                const row = (
                  <>
                    <td className="pl-4 py-3" onClick={e => e.stopPropagation()}>{file.name !== '..' && <Checkbox checked={fm.selected.has(file.name)} onChange={() => fm.toggleSelect(file.name)} />}</td>
                    <td className="px-3 py-3">
                      <div className="flex items-center gap-3">
                        <FileIcon name={file.name} is_dir={file.is_dir} />
                        <span className="text-sm text-neutral-100 truncate">{file.name}</span>
                      </div>
                    </td>
                    <td className="px-3 py-3 text-sm text-neutral-400">{file.is_dir ? '\u2014' : formatBytes(file.size)}</td>
                    <td className="px-3 py-3 text-sm text-neutral-400">{formatDate(file.mod_time)}</td>
                    <td className="px-3 py-3">{file.name !== '..' && <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>}</td>
                  </>
                );
                return actions.length > 0 ? (
                  <ContextMenuZone as="tr" key={file.name} items={actions} className={`hover:bg-neutral-800/50 cursor-pointer ${fm.selected.has(file.name) ? 'bg-neutral-800/30' : ''}`} onClick={() => fm.navigateTo(file)}>
                    {row}
                  </ContextMenuZone>
                ) : (
                  <tr key={file.name} className="hover:bg-neutral-800/50 cursor-pointer" onClick={() => fm.navigateTo(file)}>
                    {row}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <div className="md:hidden divide-y divide-neutral-800">
          {fm.loading ? (
            <div className="px-4 py-8 text-center text-sm text-neutral-500">Loading...</div>
          ) : fm.searchResults ? (
            fm.searchResults.length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-neutral-500">No results found</div>
            ) : fm.searchResults.map(result => (
              <ContextMenuZone key={result.path} items={getSearchActions(result)} className="flex items-center gap-3 px-4 py-3 hover:bg-neutral-800/50 active:bg-neutral-800/70" onClick={() => fm.navigateToSearchResult(result)}>
                <FileIcon name={result.name} is_dir={result.is_dir} />
                <div className="flex-1 min-w-0">
                  <div className="text-sm text-neutral-100 truncate">{result.name}</div>
                  <div className="text-xs text-neutral-500">{result.is_dir ? 'Folder' : formatBytes(result.size)}</div>
                </div>
                <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
              </ContextMenuZone>
            ))
          ) : fm.files.length === 0 ? (
            <div className="px-4 py-8 text-center text-sm text-neutral-500">No files found</div>
          ) : fm.files.map(file => {
            const actions = getFileActions(file);
            const content = (
              <>
                {file.name !== '..' && (
                  <div onClick={e => e.stopPropagation()}>
                    <Checkbox checked={fm.selected.has(file.name)} onChange={() => fm.toggleSelect(file.name)} />
                  </div>
                )}
                <FileIcon name={file.name} is_dir={file.is_dir} />
                <div className="flex-1 min-w-0">
                  <div className="text-sm text-neutral-100 truncate">{file.name}</div>
                  <div className="text-xs text-neutral-500">{file.is_dir ? 'Folder' : formatBytes(file.size)}</div>
                </div>
                {file.name !== '..' && (
                  <Button variant="ghost"><Icons.ellipsis className="w-5 h-5" /></Button>
                )}
              </>
            );
            return actions.length > 0 ? (
              <ContextMenuZone
                key={file.name}
                items={actions}
                className={`flex items-center gap-3 px-4 py-3 hover:bg-neutral-800/50 active:bg-neutral-800/70 ${fm.selected.has(file.name) ? 'bg-neutral-800/30' : ''}`}
                onClick={() => fm.navigateTo(file)}
              >
                {content}
              </ContextMenuZone>
            ) : (
              <div
                key={file.name}
                className={`flex items-center gap-3 px-4 py-3 hover:bg-neutral-800/50 active:bg-neutral-800/70 ${fm.selected.has(file.name) ? 'bg-neutral-800/30' : ''}`}
                onClick={() => fm.navigateTo(file)}
              >
                {content}
              </div>
            );
          })}
        </div>
      </div>

      <CreateFolderModal open={modals.newFolder} onClose={() => setModals(m => ({ ...m, newFolder: false }))} onCreate={async n => { await fm.actions.createFolder(n); setModals(m => ({ ...m, newFolder: false })); }} />
      <CreateFileModal open={modals.newFile} onClose={() => setModals(m => ({ ...m, newFile: false }))} onCreate={async n => { await fm.actions.createFile(n); setModals(m => ({ ...m, newFile: false })); }} />
      <UploadModal open={modals.upload} onClose={() => { setModals(m => ({ ...m, upload: false, initialFiles: null })); if (fm.uploadInputRef.current) fm.uploadInputRef.current.value = ''; }} serverId={id || ''} path={fm.currentPath} onComplete={fm.refreshFiles} initialFiles={modals.initialFiles} />
      <MoveFileModal open={fileTarget?.type === 'move'} initialPath={fileTarget?.type === 'move' ? fm.getFilePath(fileTarget.file.name) : ''} onClose={() => setFileTarget(null)} onMove={async dest => { if (fileTarget?.type === 'move') await fm.actions.move(fileTarget.file, dest); setFileTarget(null); }} />
      <RenameFileModal open={fileTarget?.type === 'rename'} initialName={fileTarget?.file.name || ''} isDir={fileTarget?.file.is_dir || false} onClose={() => setFileTarget(null)} onRename={async n => { if (fileTarget?.type === 'rename') await fm.actions.rename(fileTarget.file, n); setFileTarget(null); }} />
      <CompressFileModal open={fileTarget?.type === 'compress'} fileName={fileTarget?.file.name || ''} onClose={() => setFileTarget(null)} onCompress={async fmt => { if (fileTarget?.type === 'compress') await fm.actions.compress(fileTarget.file, fmt); setFileTarget(null); }} />
      <CompressFileModal open={modals.bulkCompress} fileName={`${fm.selected.size} item${fm.selected.size > 1 ? 's' : ''}`} onClose={() => setModals(m => ({ ...m, bulkCompress: false }))} onCompress={async fmt => { await fm.actions.bulkCompress(fmt); setModals(m => ({ ...m, bulkCompress: false })); }} />

      <DeleteFileModal
        open={!!deleteTarget}
        fileName={deleteTarget && 'file' in deleteTarget ? deleteTarget.file.name : ''}
        isDir={deleteTarget && 'file' in deleteTarget ? deleteTarget.file.is_dir : false}
        isBulk={!!(deleteTarget && 'bulk' in deleteTarget)}
        count={fm.selected.size}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          if (deleteTarget && 'file' in deleteTarget) {
            await fm.actions.delete(deleteTarget.file);
          } else if (deleteTarget && 'bulk' in deleteTarget) {
            await fm.actions.bulkDelete();
          }
        }}
      />

      <ClipboardPanel items={fm.clipboard} pasting={fm.pasting} onPaste={fm.actions.paste} onClear={() => fm.setClipboard([])} onRemove={p => fm.setClipboard(fm.clipboard.filter(x => x !== p))} />

      <BulkActionBar count={fm.selected.size} onClear={() => fm.toggleAll()}>
        <Button variant="secondary" onClick={fm.actions.bulkCopy} className="px-2 sm:px-3"><Icons.copy className="h-4 w-4" /><span className="hidden sm:inline ml-1">Copy</span></Button>
        <Button variant="secondary" onClick={() => setModals(m => ({ ...m, bulkCompress: true }))} className="px-2 sm:px-3"><Icons.archive className="h-4 w-4" /><span className="hidden sm:inline ml-1">Compress</span></Button>
        <Button variant="danger" onClick={() => setDeleteTarget({ bulk: true })} className="px-2 sm:px-3"><Icons.trash className="h-4 w-4" /><span className="hidden sm:inline ml-1">Delete</span></Button>
      </BulkActionBar>
    </div>
  );
}
