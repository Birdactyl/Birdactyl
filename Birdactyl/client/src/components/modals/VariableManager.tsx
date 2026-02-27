import { useState } from 'react';
import { PackageVariable } from '../../lib/api';
import { Input, Checkbox, Icons } from '../';

interface Props {
  variables: PackageVariable[];
  onChange: (variables: PackageVariable[]) => void;
}

export default function VariableManager({ variables, onChange }: Props) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [defaultValue, setDefaultValue] = useState('');
  const [userEditable, setUserEditable] = useState(true);
  const [rules, setRules] = useState('');
  const [editing, setEditing] = useState(-1);

  const addOrUpdate = () => {
    if (!name) return;
    const newVar = { name, description, default: defaultValue, user_editable: userEditable, rules };
    if (editing >= 0) {
      onChange(variables.map((v, i) => i === editing ? newVar : v));
      setEditing(-1);
    } else {
      onChange([...variables, newVar]);
    }
    setName('');
    setDescription('');
    setDefaultValue('');
    setUserEditable(true);
    setRules('');
  };

  const edit = (index: number) => {
    const v = variables[index];
    setName(v.name);
    setDescription(v.description);
    setDefaultValue(v.default);
    setUserEditable(v.user_editable);
    setRules(v.rules || '');
    setEditing(index);
  };

  const remove = (index: number) => {
    onChange(variables.filter((_, i) => i !== index));
    if (editing === index) setEditing(-1);
  };

  return (
    <div className="space-y-4">
      <div className="p-4 rounded-lg bg-neutral-900/50 border border-neutral-800 space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <Input label="Variable Name" value={name} onChange={e => setName(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, ''))} placeholder="SERVER_MEMORY" />
          <Input label="Default Value" value={defaultValue} onChange={e => setDefaultValue(e.target.value)} placeholder="1024" />
        </div>
        <Input label="Description" value={description} onChange={e => setDescription(e.target.value)} placeholder="Server memory allocation in MB" />
        <div className="grid grid-cols-2 gap-3">
          <Input label="Validation Rules" value={rules} onChange={e => setRules(e.target.value)} placeholder="required|numeric|min:512" />
          <div className="flex items-end pb-2">
            <Checkbox checked={userEditable} onChange={(c) => setUserEditable(c ?? false)} label="User Editable" />
          </div>
        </div>
        <button
          type="button"
          onClick={addOrUpdate}
          disabled={!name}
          className="w-full py-2 text-sm font-medium text-neutral-300 bg-neutral-800 hover:bg-neutral-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {editing >= 0 ? 'Update Variable' : 'Add Variable'}
        </button>
      </div>

      {variables.length > 0 ? (
        <div className="space-y-2">
          <div className="text-xs font-medium text-neutral-400">Added Variables</div>
          {variables.map((v, i) => (
            <div key={i} className={`flex items-center justify-between p-3 rounded-lg ${editing === i ? 'bg-amber-500/20 ring-1 ring-amber-500' : 'bg-neutral-800/50'}`}>
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-mono font-medium text-neutral-100">{v.name}</span>
                  {v.user_editable && <span className="text-xs text-emerald-400 bg-emerald-500/20 px-2 py-0.5 rounded">Editable</span>}
                </div>
                {v.description && <p className="text-xs text-neutral-400 mt-0.5">{v.description}</p>}
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
        <p className="text-sm text-neutral-500 text-center py-4">No variables added yet. Variables allow customization of the server.</p>
      )}
    </div>
  );
}
