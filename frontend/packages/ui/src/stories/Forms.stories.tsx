import type { Meta, StoryObj } from '@storybook/react-vite';

import { Button, Checkbox, Field, Input, Label, Select, Textarea } from '../index.js';

const meta = {
  title: 'Components/Forms',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const Controls: Story = {
  render: () => (
    <form
      className="mercadia-story-panel"
      style={{ display: 'grid', gap: 'var(--ui-space-lg)', width: 420 }}
    >
      <Field>
        <Label htmlFor="cashier" hint="Required for audit trail">
          Cashier ID
        </Label>
        <Input id="cashier" placeholder="cashier-001" />
      </Field>
      <Field>
        <Label htmlFor="operation">Operation</Label>
        <Select id="operation" defaultValue="cash-in">
          <option value="cash-in">Cash in</option>
          <option value="cash-out">Cash out</option>
          <option value="recount">Safe recount</option>
        </Select>
      </Field>
      <Field>
        <Label htmlFor="reason">Reason</Label>
        <Textarea id="reason" placeholder="Describe the operation reason" rows={4} />
      </Field>
      <label className="mercadia-story-row">
        <Checkbox />
        <span>Supervisor approval captured</span>
      </label>
      <div className="mercadia-story-row">
        <Button>Submit</Button>
        <Button variant="secondary">Cancel</Button>
      </div>
    </form>
  ),
};
