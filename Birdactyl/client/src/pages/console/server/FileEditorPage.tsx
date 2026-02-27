import { useEffect, useState } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import Editor from '@monaco-editor/react';
import { getServer, Server, readFile, writeFile } from '../../../lib/api';
import { Button, PermissionDenied } from '../../../components';
import { notify } from '../../../components/feedback/Notification';

function getLanguage(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  const map: Record<string, string> = {
    js: 'javascript', ts: 'typescript', jsx: 'javascript', tsx: 'typescript',
    json: 'json', yaml: 'yaml', yml: 'yaml', xml: 'xml', html: 'html',
    css: 'css', scss: 'scss', md: 'markdown', py: 'python', go: 'go',
    rs: 'rust', java: 'java', sh: 'shell', bash: 'shell', properties: 'ini',
    toml: 'ini', cfg: 'ini', conf: 'ini', txt: 'plaintext',
  };
  return map[ext] || 'plaintext';
}

export default function FileEditorPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const filePath = searchParams.get('path') || '/';
  const fileName = filePath.split('/').pop() || 'file';
  const dirPath = filePath.substring(0, filePath.lastIndexOf('/')) || '/';

  const [server, setServer] = useState<Server | null>(null);
  const [content, setContent] = useState('');
  const [originalContent, setOriginalContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const hasChanges = content !== originalContent;

  useEffect(() => {
    if (!id) return;
    getServer(id).then(res => {
      if (res.success && res.data) setServer(res.data);
    });
  }, [id]);

  useEffect(() => {
    if (!id || !filePath) return;
    setLoading(true);
    readFile(id, filePath).then(res => {
      if (res.success && res.data !== undefined) {
        setContent(res.data);
        setOriginalContent(res.data);
      } else if (res.error === 'Permission denied') {
        setError("You don't have permission to read this file");
      }
      setLoading(false);
    });
  }, [id, filePath]);

  const handleSave = async () => {
    if (!id || saving || !hasChanges) return;
    setSaving(true);
    const res = await writeFile(id, filePath, content);
    if (res.success) {
      setOriginalContent(content);
    } else if (res.error) {
      notify('Error', res.error.includes('ermission') ? "You don't have permission to edit this file" : res.error, 'error');
    }
    setSaving(false);
  };

  const goBack = () => {
    navigate(`/console/server/${id}/files${dirPath !== '/' ? `?path=${encodeURIComponent(dirPath)}` : ''}`);
  };

  const pathParts = filePath.split('/').filter(Boolean);

  return (
    <div className="space-y-6">
      {error && <PermissionDenied message={error} />}

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-1 text-sm text-neutral-600 dark:text-neutral-400">
          <span className="font-medium text-neutral-800 dark:text-neutral-200">{server?.name || 'Server'}</span>
          <span className="text-neutral-400">/</span>
          <button type="button" className="hover:underline decoration-dotted" onClick={() => navigate(`/console/server/${id}/files`)}>Files</button>
          {pathParts.slice(0, -1).map((part, i) => (
            <span key={i} className="flex items-center gap-1">
              <span className="text-neutral-400">/</span>
              <button
                type="button"
                className="hover:underline decoration-dotted text-neutral-700 dark:text-neutral-300"
                onClick={() => navigate(`/console/server/${id}/files?path=${encodeURIComponent('/' + pathParts.slice(0, i + 1).join('/'))}`)}
              >
                {part}
              </button>
            </span>
          ))}
          <span className="text-neutral-400">/</span>
          <span className="font-semibold text-neutral-900 dark:text-neutral-100">{fileName}</span>
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
          <Button variant="ghost" onClick={goBack}>Back</Button>
          <Button onClick={handleSave} disabled={!hasChanges} loading={saving}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-900 dark:text-neutral-100">Editing {fileName}</h1>
          <p className="text-sm text-neutral-600 dark:text-neutral-400">{filePath} · {getLanguage(fileName)}</p>
        </div>
      </div>

      <div className="border border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden">
        {loading ? (
          <div className="h-[70vh] flex items-center justify-center text-neutral-400">Loading...</div>
        ) : (
          <Editor
            height="70vh"
            language={getLanguage(fileName)}
            value={content}
            onChange={v => setContent(v || '')}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              scrollBeyondLastLine: false,
              wordWrap: 'on',
              automaticLayout: true,
            }}
          />
        )}
      </div>
    </div>
  );
}
