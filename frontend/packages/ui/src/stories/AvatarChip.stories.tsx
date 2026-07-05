import type { Meta, StoryObj } from '@storybook/react-vite';

import { AvatarChip, ThemePreview } from '../index.js';

const meta = {
  title: 'Components/AvatarChip',
  component: AvatarChip,
  parameters: {
    layout: 'centered',
  },
  argTypes: {
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
  },
  args: {
    initials: 'MB',
    size: 'md',
  },
} satisfies Meta<typeof AvatarChip>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="mercadia-story-row">
      <AvatarChip initials="MB" size="sm" />
      <AvatarChip initials="MB" size="md" />
      <AvatarChip initials="MB" size="lg" />
    </div>
  ),
};

export const DerivedPalette: Story = {
  render: () => (
    <div className="mercadia-story-row">
      <AvatarChip initials="MB" />
      <AvatarChip initials="СИ" />
      <AvatarChip initials="АЗ" />
      <AvatarChip initials="ДР" />
    </div>
  ),
};

export const ExplicitColorOverride: Story = {
  render: () => <AvatarChip initials="XY" color="#7C3AED" />,
};

/**
 * Renders the derived avatar-chip palette inside the senior-cashier dark
 * surface (Plan 018 palette: `--ui-bg: #0d0a07`, `--ui-surface: #1a140e`,
 * accent `#EA580C`), matching how the design's top bar and lists show
 * multiple people's avatar chips together.
 */
export const OnSeniorCashierSurface: Story = {
  render: () => (
    <ThemePreview
      className="mercadia-theme-sample"
      theme={{ surface: 'senior-cashier', colorMode: 'dark', accentPreset: 'senior-cashier' }}
    >
      <div className="mercadia-story-row">
        <AvatarChip initials="MB" />
        <AvatarChip initials="СИ" />
        <AvatarChip initials="АЗ" />
        <AvatarChip initials="ДР" size="lg" />
      </div>
    </ThemePreview>
  ),
};
