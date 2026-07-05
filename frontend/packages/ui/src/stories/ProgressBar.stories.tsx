import type { Meta, StoryObj } from '@storybook/react-vite';

import { ProgressBar, ThemePreview } from '../index.js';

const meta = {
  title: 'Components/ProgressBar',
  component: ProgressBar,
  parameters: {
    layout: 'centered',
  },
  argTypes: {
    value: {
      control: { type: 'range', min: 0, max: 100, step: 1 },
    },
  },
  args: {
    value: 45,
  },
} satisfies Meta<typeof ProgressBar>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => <ProgressBar {...args} style={{ width: '320px' }} />,
};

export const ValueSteps: Story = {
  render: () => (
    <div className="mercadia-story-section" style={{ width: '320px' }}>
      <ProgressBar value={0} />
      <ProgressBar value={25} />
      <ProgressBar value={45} />
      <ProgressBar value={75} />
      <ProgressBar value={100} />
    </div>
  ),
};

/**
 * Renders the progress bar inside the senior-cashier dark surface (Plan 018
 * palette). Note: this is the flat single-accent-fill simplification, not
 * the design's green-to-orange gradient fill — see ProgressBar.tsx.
 */
export const OnSeniorCashierSurface: Story = {
  render: () => (
    <ThemePreview
      className="mercadia-theme-sample"
      theme={{ surface: 'senior-cashier', colorMode: 'dark', accentPreset: 'senior-cashier' }}
    >
      <div style={{ width: '320px' }}>
        <ProgressBar value={68} />
      </div>
    </ThemePreview>
  ),
};
