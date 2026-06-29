import type { Meta, StoryObj } from '@storybook/react-vite';

import { Button, Checkbox, Field, Input, Label, Select, Textarea } from '../index.js';
import { getStorybookLocale } from './locale.js';

const formCopy = {
  en: {
    cashierHint: 'Required for audit trail',
    cashierId: 'Cashier ID',
    operation: 'Operation',
    cashIn: 'Cash in',
    cashOut: 'Cash out',
    safeRecount: 'Safe recount',
    reason: 'Reason',
    reasonPlaceholder: 'Describe the operation reason',
    supervisorCaptured: 'Supervisor approval captured',
    submit: 'Submit',
    cancel: 'Cancel',
  },
  ru: {
    cashierHint: 'Обязательно для аудита',
    cashierId: 'ID кассира',
    operation: 'Операция',
    cashIn: 'Внесение наличных',
    cashOut: 'Изъятие наличных',
    safeRecount: 'Пересчет сейфа',
    reason: 'Причина',
    reasonPlaceholder: 'Опишите причину операции',
    supervisorCaptured: 'Подтверждение supervisor получено',
    submit: 'Отправить',
    cancel: 'Отмена',
  },
};

const meta = {
  title: 'Components/Forms',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const Controls: Story = {
  render: (_args, context) => {
    const copy = formCopy[getStorybookLocale(context.globals.locale)];
    return (
      <form
        className="mercadia-story-panel"
        style={{ display: 'grid', gap: 'var(--ui-space-lg)', width: 420 }}
      >
        <Field>
          <Label htmlFor="cashier" hint={copy.cashierHint}>
            {copy.cashierId}
          </Label>
          <Input id="cashier" placeholder="cashier-001" />
        </Field>
        <Field>
          <Label htmlFor="operation">{copy.operation}</Label>
          <Select id="operation" defaultValue="cash-in">
            <option value="cash-in">{copy.cashIn}</option>
            <option value="cash-out">{copy.cashOut}</option>
            <option value="recount">{copy.safeRecount}</option>
          </Select>
        </Field>
        <Field>
          <Label htmlFor="reason">{copy.reason}</Label>
          <Textarea id="reason" placeholder={copy.reasonPlaceholder} rows={4} />
        </Field>
        <label className="mercadia-story-row">
          <Checkbox />
          <span>{copy.supervisorCaptured}</span>
        </label>
        <div className="mercadia-story-row">
          <Button>{copy.submit}</Button>
          <Button variant="secondary">{copy.cancel}</Button>
        </div>
      </form>
    );
  },
};
