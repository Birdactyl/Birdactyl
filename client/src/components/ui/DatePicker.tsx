import { useState, useRef, useEffect } from 'react';
import { Icons } from '../Icons';

interface DatePickerProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  className?: string;
}

const DAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

function TimeScroller({ value, max, onChange }: { value: number; max: number; onChange: (v: number) => void }) {
  const ref = useRef<HTMLDivElement>(null);

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    if (e.deltaY > 0) onChange((value + 1) % max);
    else onChange((value - 1 + max) % max);
  };

  return (
    <div ref={ref} onWheel={handleWheel} className="flex flex-col items-center select-none cursor-ns-resize">
      {[-2, -1, 0, 1, 2].map(offset => {
        const num = (value + offset + max) % max;
        const isCenter = offset === 0;
        return (
          <button
            key={offset}
            type="button"
            onClick={() => onChange(num)}
            className={`w-full py-1 text-xs text-center transition ${isCenter ? 'text-neutral-100 font-medium text-sm' : 'text-neutral-600'}`}
          >
            {String(num).padStart(2, '0')}
          </button>
        );
      })}
    </div>
  );
}

export default function DatePicker({ label, value, onChange, className = '' }: DatePickerProps) {
  const [open, setOpen] = useState(false);
  const [viewDate, setViewDate] = useState(() => value ? new Date(value) : new Date());
  const [hour, setHour] = useState(() => value ? new Date(value).getHours() : 0);
  const [minute, setMinute] = useState(() => value ? new Date(value).getMinutes() : 0);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  useEffect(() => {
    if (value) {
      const d = new Date(value);
      setViewDate(d);
      setHour(d.getHours());
      setMinute(d.getMinutes());
    }
  }, [value]);

  const selectedDate = value ? new Date(value) : null;
  const year = viewDate.getFullYear();
  const month = viewDate.getMonth();
  const firstDay = new Date(year, month, 1).getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const today = new Date();

  const days: (number | null)[] = [];
  for (let i = 0; i < firstDay; i++) days.push(null);
  for (let i = 1; i <= daysInMonth; i++) days.push(i);

  const selectDay = (day: number) => {
    const d = new Date(year, month, day, hour, minute);
    onChange(d.toISOString());
  };

  const prevMonth = () => setViewDate(new Date(year, month - 1, 1));
  const nextMonth = () => setViewDate(new Date(year, month + 1, 1));

  const updateTime = (h: number, m: number) => {
    setHour(h);
    setMinute(m);
    if (selectedDate) {
      const d = new Date(selectedDate);
      d.setHours(h, m);
      onChange(d.toISOString());
    }
  };

  const clear = () => {
    onChange('');
    setOpen(false);
  };

  const formatDisplay = () => {
    if (!selectedDate) return '';
    return `${selectedDate.getMonth() + 1}/${selectedDate.getDate()}/${selectedDate.getFullYear()} ${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`;
  };

  const isSelected = (day: number) => selectedDate && selectedDate.getFullYear() === year && selectedDate.getMonth() === month && selectedDate.getDate() === day;
  const isToday = (day: number) => today.getFullYear() === year && today.getMonth() === month && today.getDate() === day;

  return (
    <div className={`flex items-center gap-2 ${className}`} ref={ref}>
      {label && <span className="text-xs text-neutral-400">{label}</span>}
      <div className="relative">
        <button
          type="button"
          onClick={() => setOpen(!open)}
          className="flex items-center gap-2 rounded-lg border border-neutral-800/60 bg-neutral-900/60 text-neutral-100 transition hover:border-neutral-500 focus:outline-none focus:ring-2 focus:ring-neutral-100 focus:ring-offset-2 focus:ring-offset-[#0a0a0a] px-3 py-2 text-xs min-w-[180px]"
        >
          <Icons.calendar className="w-4 h-4 text-neutral-400" />
          <span className={value ? 'text-neutral-100' : 'text-neutral-500'}>{value ? formatDisplay() : 'Select date...'}</span>
        </button>

        {open && (
          <div className="absolute top-full left-0 mt-2 z-50 rounded-lg border border-neutral-800 bg-neutral-900 shadow-xl flex">
            <div className="p-3 w-[240px]">
              <div className="flex items-center justify-between mb-3">
                <button type="button" onClick={prevMonth} className="p-1 hover:bg-neutral-800 rounded transition">
                  <Icons.chevronLeft className="w-4 h-4 text-neutral-400" />
                </button>
                <span className="text-sm font-medium text-neutral-100">{MONTHS[month]} {year}</span>
                <button type="button" onClick={nextMonth} className="p-1 hover:bg-neutral-800 rounded transition">
                  <Icons.chevronRight className="w-4 h-4 text-neutral-400" />
                </button>
              </div>

              <div className="grid grid-cols-7 gap-1 mb-1">
                {DAYS.map(d => <div key={d} className="text-center text-[10px] font-medium text-neutral-500 py-1">{d}</div>)}
              </div>

              <div className="grid grid-cols-7 gap-1">
                {days.map((day, i) => (
                  <button
                    key={i}
                    type="button"
                    disabled={!day}
                    onClick={() => day && selectDay(day)}
                    className={`h-7 w-7 rounded text-xs transition ${!day ? '' : isSelected(day) ? 'bg-neutral-100 text-neutral-900 font-medium' : isToday(day) ? 'bg-neutral-800 text-neutral-100' : 'text-neutral-300 hover:bg-neutral-800'}`}
                  >
                    {day}
                  </button>
                ))}
              </div>

              <div className="mt-3 pt-3 border-t border-neutral-800 flex justify-end">
                <button type="button" onClick={clear} className="text-xs text-neutral-400 hover:text-neutral-200 transition">Clear</button>
              </div>
            </div>

            <div className="border-l border-neutral-800 flex">
              <div className="w-12 flex flex-col justify-center bg-neutral-800/30">
                <TimeScroller value={hour} max={24} onChange={h => updateTime(h, minute)} />
              </div>
              <div className="w-12 flex flex-col justify-center bg-neutral-800/30">
                <TimeScroller value={minute} max={60} onChange={m => updateTime(hour, m)} />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
