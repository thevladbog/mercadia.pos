import { cn } from '../../lib/cn.js';

type NumpadProps = {
  value: string;
  onChange: (value: string) => void;
  onEnter?: () => void;
  className?: string;
  enterLabel?: string;
  /**
   * When true, the built-in display renders one `•` per entered character
   * instead of the raw value — for PIN-style entry where the digits must
   * never appear in plaintext (plan 020). Defaults to `false`, which
   * preserves `pos-terminal`'s existing money-amount-entry usage
   * byte-for-byte (it never passes this prop).
   */
  mask?: boolean;
};

const digitKeys = ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as const;

export function Numpad({
  value,
  onChange,
  onEnter,
  className,
  enterLabel = 'OK',
  mask = false,
}: NumpadProps) {
  function append(digit: string) {
    onChange(`${value}${digit}`);
  }

  function backspace() {
    onChange(value.slice(0, -1));
  }

  const displayValue = mask ? '•'.repeat(value.length) : value;

  return (
    <div className={cn('mercadia-numpad', className)}>
      <div
        className="mercadia-numpad-display"
        aria-label={mask ? `${value.length} digits entered` : undefined}
      >
        {displayValue || '0'}
      </div>
      {digitKeys.map((digit) => (
        <button
          key={digit}
          className="mercadia-numpad-key"
          onClick={() => append(digit)}
          type="button"
        >
          {digit}
        </button>
      ))}
      <button className="mercadia-numpad-key" onClick={backspace} type="button">
        ⌫
      </button>
      <button className="mercadia-numpad-key" onClick={() => append('0')} type="button">
        0
      </button>
      <button
        className="mercadia-numpad-key mercadia-numpad-key--accent"
        onClick={onEnter}
        type="button"
      >
        {enterLabel}
      </button>
    </div>
  );
}
