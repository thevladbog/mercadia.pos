import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';

import { LayoutGrid, Numpad, Stepper } from '../index.js';
import type { LayoutGridSpec } from '../index.js';

const sampleGrid: LayoutGridSpec = {
  rows: 3,
  cols: 4,
  tiles: [
    { label: 'Coffee', productId: 'sku-coffee' },
    { label: 'Bakery', productId: 'sku-bakery', color: '#f97316' },
    { label: 'Water', productId: 'sku-water', color: '#2563eb' },
    { label: 'Fruit', productId: 'sku-fruit' },
    { label: 'Discount', productId: 'discount', color: '#22c55e' },
    { label: 'Return', productId: 'return', color: '#3b82f6' },
  ],
};

const meta = {
  title: 'Components/Touch controls',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const NumpadControl: Story = {
  render: function Render() {
    const [value, setValue] = useState('1250');
    return <Numpad enterLabel="Pay" value={value} onChange={setValue} />;
  },
};

export const StepperControl: Story = {
  render: function Render() {
    const [value, setValue] = useState(2);
    return <Stepper value={value} min={0} max={10} onChange={setValue} />;
  },
};

export const LayoutGridControl: Story = {
  render: () => (
    <div style={{ width: 560 }}>
      <LayoutGrid grid={sampleGrid} onTileClick={() => undefined} />
    </div>
  ),
};
