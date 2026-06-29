import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';

import { LayoutGrid, Numpad, Stepper } from '../index.js';
import type { LayoutGridSpec } from '../index.js';
import { getStorybookLocale } from './locale.js';

const touchCopy = {
  en: {
    pay: 'Pay',
    grid: ['Coffee', 'Bakery', 'Water', 'Fruit', 'Discount', 'Return'],
  },
  ru: {
    pay: 'Оплатить',
    grid: ['Кофе', 'Выпечка', 'Вода', 'Фрукты', 'Скидка', 'Возврат'],
  },
};

function sampleGrid(labels: string[]): LayoutGridSpec {
  return {
    rows: 3,
    cols: 4,
    tiles: [
      { label: labels[0] ?? '', productId: 'sku-coffee' },
      { label: labels[1] ?? '', productId: 'sku-bakery', color: '#f97316' },
      { label: labels[2] ?? '', productId: 'sku-water', color: '#2563eb' },
      { label: labels[3] ?? '', productId: 'sku-fruit' },
      { label: labels[4] ?? '', productId: 'discount', color: '#22c55e' },
      { label: labels[5] ?? '', productId: 'return', color: '#3b82f6' },
    ],
  };
}

const meta = {
  title: 'Components/Touch controls',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const NumpadControl: Story = {
  render: function Render(_args, context) {
    const [value, setValue] = useState('1250');
    const copy = touchCopy[getStorybookLocale(context.globals.locale)];
    return <Numpad enterLabel={copy.pay} value={value} onChange={setValue} />;
  },
};

export const StepperControl: Story = {
  render: function Render() {
    const [value, setValue] = useState(2);
    return <Stepper value={value} min={0} max={10} onChange={setValue} />;
  },
};

export const LayoutGridControl: Story = {
  render: (_args, context) => {
    const copy = touchCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div style={{ width: 560 }}>
        <LayoutGrid grid={sampleGrid(copy.grid)} onTileClick={() => undefined} />
      </div>
    );
  },
};
