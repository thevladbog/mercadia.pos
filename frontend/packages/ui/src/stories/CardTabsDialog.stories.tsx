import type { Meta, StoryObj } from '@storybook/react-vite';

import {
  Button,
  Card,
  CardHeading,
  Dialog,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogTrigger,
  PillTabs,
} from '../index.js';

const meta = {
  title: 'Components/Containers',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const Cards: Story = {
  render: () => (
    <div className="mercadia-story-grid" style={{ width: 760 }}>
      <Card>
        <CardHeading title="Sales" subtitle="Current operational day" />
        <strong>128 receipts</strong>
      </Card>
      <Card variant="success">
        <CardHeading title="Payments" subtitle="Settled successfully" />
        <strong>98.4%</strong>
      </Card>
      <Card variant="warning">
        <CardHeading title="Fiscal queue" subtitle="Needs operator attention" />
        <strong>3 pending</strong>
      </Card>
    </div>
  ),
};

export const Tabs: Story = {
  render: function Render() {
    return (
      <div className="mercadia-story-panel" style={{ width: 640 }}>
        <PillTabs
          aria-label="Receipt sections"
          items={[
            {
              value: 'lines',
              label: 'Lines',
              count: 12,
              content: 'Receipt line table placeholder',
            },
            {
              value: 'payments',
              label: 'Payments',
              count: 2,
              content: 'Payment state placeholder',
            },
            { value: 'fiscal', label: 'Fiscal', count: 1, content: 'Fiscal document placeholder' },
          ]}
          onValueChange={() => undefined}
          value="lines"
        />
      </div>
    );
  },
};

export const Dialogs: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Open dialog</Button>
      </DialogTrigger>
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>Close shift</DialogTitle>
        <DialogBody>
          Verify cash balance, open payments, and fiscal documents before closing this shift.
        </DialogBody>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">Cancel</Button>
          </DialogClose>
          <Button>Close shift</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};
