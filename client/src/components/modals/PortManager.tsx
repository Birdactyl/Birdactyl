import { useState } from 'react';
import { PackagePort } from '../../lib/api';
import { Input, Checkbox, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, Icons } from '../';

interface Props {
  ports: PackagePort[];
  onChange: (ports: PackagePort[]) => void;
}

export default function PortManager({ ports, onChange }: Props) {
  const [name, setName] = useState('');
  const [defaultPort, setDefaultPort] = useState('');
  const [protocol, setProtocol] = useState('tcp');
  const [primary, setPrimary] = useState(false);
  const [editing, setEditing] = useState(-1);

  const addOrUpdate = () => {
    if (!name || !defaultPort) return;
    const newPort = { name, default: parseInt(defaultPort), protocol, primary };
    if (editing >= 0) {
      onChange(ports.map((p, i) => i === editing ? newPort : p));
      setEditing(-1);
    } else {
      onChange([...ports, newPort]);
    }
    setName('');
    setDefaultPort('');
    setProtocol('tcp');
    setPrimary(false);
  };

  const edit = (index: number) => {
    const p = ports[index];
    setName(p.name);
    setDefaultPort(String(p.default));
    setProtocol(p.protocol);
    setPrimary(p.primary || false);
    setEditing(index);
  };

  const remove = (index: number) => {
    onChange(ports.filter((_, i) => i !== index));
    if (editing === index) setEditing(-1);
  };

  return (
    <div className="space-y-4">
      <div className="p-4 rounded-lg bg-neutral-900/50 border border-neutral-800 space-y-3">
        <div className="grid grid-cols-4 gap-3">
          <Input label="Port Name" value={name} onChange={e => setName(e.target.value)} placeholder="Game" />
          <Input label="Default Port" type="number" value={defaultPort} onChange={e => setDefaultPort(e.target.value)} placeholder="25565" />
          <div className="flex flex-col gap-1.5">
            <label className="block text-xs font-medium text-neutral-400">Protocol</label>
            <DropdownMenu className="w-full">
              <DropdownMenuTrigger asChild className="w-full">
                <button type="button" className="w-full rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 text-left transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-[#0a0a0a] px-3 py-2 flex items-center justify-between" style={{ fontSize: '13px' }}>
                  {protocol.toUpperCase()}
                  <Icons.chevronDown className="w-4 h-4 text-neutral-400" />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onSelect={() => setProtocol('tcp')}>TCP</DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setProtocol('udp')}>UDP</DropdownMenuItem>
                <DropdownMenuItem onSelect={() => setProtocol('both')}>Both</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <div className="flex items-end pb-2">
            <Checkbox checked={primary} onChange={(c) => setPrimary(c ?? false)} label="Primary" />
          </div>
        </div>
        <button
          type="button"
          onClick={addOrUpdate}
          disabled={!name || !defaultPort}
          className="w-full py-2 text-sm font-medium text-neutral-300 bg-neutral-800 hover:bg-neutral-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {editing >= 0 ? 'Update Port' : 'Add Port'}
        </button>
      </div>

      {ports.length > 0 ? (
        <div className="space-y-2">
          <div className="text-xs font-medium text-neutral-400">Added Ports</div>
          {ports.map((port, i) => (
            <div key={i} className={`flex items-center justify-between p-3 rounded-lg ${editing === i ? 'bg-amber-500/20 ring-1 ring-amber-500' : 'bg-neutral-800/50'}`}>
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-neutral-100">{port.name}</span>
                <span className="text-sm text-neutral-400">{port.default}/{port.protocol}</span>
                {port.primary && <span className="text-xs text-amber-400 bg-amber-500/20 px-2 py-0.5 rounded">Primary</span>}
              </div>
              <div className="flex items-center gap-1">
                <button onClick={() => edit(i)} className="text-neutral-400 hover:text-neutral-100 transition-colors p-1">
                  <Icons.edit className="w-4 h-4" />
                </button>
                <button onClick={() => remove(i)} className="text-neutral-400 hover:text-red-400 transition-colors p-1">
                  <Icons.x className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-sm text-neutral-500 text-center py-4">No ports added yet. Add at least one port for the server.</p>
      )}
    </div>
  );
}
