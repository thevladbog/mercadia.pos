import { Checkbox, Field, Input, Label, Select } from '@mercadia/ui';
import { useId, type ComponentPropsWithoutRef, type ReactNode } from 'react';

type TextFieldProps = Omit<ComponentPropsWithoutRef<typeof Input>, 'onChange'> & {
  label: ReactNode;
  onValueChange: (value: string) => void;
};

export function TextField({ id, label, onValueChange, ...props }: TextFieldProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;

  return (
    <Field>
      <Label htmlFor={inputId}>{label}</Label>
      <Input id={inputId} onChange={(event) => onValueChange(event.target.value)} {...props} />
    </Field>
  );
}

type SelectFieldProps = Omit<ComponentPropsWithoutRef<typeof Select>, 'onChange'> & {
  label: ReactNode;
  onValueChange: (value: string) => void;
};

export function SelectField({ id, label, onValueChange, children, ...props }: SelectFieldProps) {
  const generatedId = useId();
  const selectId = id ?? generatedId;

  return (
    <Field>
      <Label htmlFor={selectId}>{label}</Label>
      <Select id={selectId} onChange={(event) => onValueChange(event.target.value)} {...props}>
        {children}
      </Select>
    </Field>
  );
}

type CheckboxFieldProps = Omit<
  ComponentPropsWithoutRef<typeof Checkbox>,
  'className' | 'onChange' | 'type'
> & {
  className?: string;
  label: ReactNode;
  onCheckedChange: (checked: boolean) => void;
};

export function CheckboxField({
  className,
  id,
  label,
  onCheckedChange,
  ...props
}: CheckboxFieldProps) {
  const generatedId = useId();
  const checkboxId = id ?? generatedId;

  return (
    <label className={['checkbox-field', className].filter(Boolean).join(' ')} htmlFor={checkboxId}>
      <Checkbox
        id={checkboxId}
        onChange={(event) => onCheckedChange(event.target.checked)}
        {...props}
      />
      <span>{label}</span>
    </label>
  );
}
