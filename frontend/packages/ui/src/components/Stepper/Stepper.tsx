import { cn } from '../../lib/cn.js';

type StepperProps = {
  value: number;
  min?: number;
  max?: number;
  step?: number;
  onChange: (value: number) => void;
  className?: string;
};

export function Stepper({
  value,
  min = 0,
  max = 999,
  step = 1,
  onChange,
  className,
}: StepperProps) {
  function decrement() {
    onChange(Math.max(min, value - step));
  }

  function increment() {
    onChange(Math.min(max, value + step));
  }

  return (
    <div className={cn('mercadia-stepper', className)}>
      <button
        className="mercadia-stepper-button"
        disabled={value <= min}
        onClick={decrement}
        type="button"
      >
        −
      </button>
      <span className="mercadia-stepper-value">{value}</span>
      <button
        className="mercadia-stepper-button mercadia-stepper-button--primary"
        disabled={value >= max}
        onClick={increment}
        type="button"
      >
        +
      </button>
    </div>
  );
}
