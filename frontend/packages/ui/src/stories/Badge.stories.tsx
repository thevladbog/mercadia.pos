import type { Meta, StoryObj } from '@storybook/react-vite';

import { Badge } from '../index.js';
import { getStorybookLocale } from './locale.js';

const badgeCopy = {
  en: ['Default', 'Accent', 'Success', 'Warning', 'Danger', 'Info', 'Outline'],
  ru: ['Базовый', 'Акцент', 'Успех', 'Внимание', 'Опасность', 'Инфо', 'Контур'],
};

const meta = {
  title: 'Components/Badge',
  component: Badge,
  parameters: {
    layout: 'centered',
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'accent', 'success', 'warning', 'danger', 'info', 'outline'],
    },
  },
  args: {
    children: 'Open shift',
    variant: 'accent',
  },
} satisfies Meta<typeof Badge>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Playground: Story = {};

export const Variants: Story = {
  render: (_args, context) => {
    const copy = badgeCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div className="mercadia-story-row">
        <Badge>{copy[0]}</Badge>
        <Badge variant="accent">{copy[1]}</Badge>
        <Badge variant="success">{copy[2]}</Badge>
        <Badge variant="warning">{copy[3]}</Badge>
        <Badge variant="danger">{copy[4]}</Badge>
        <Badge variant="info">{copy[5]}</Badge>
        <Badge variant="outline">{copy[6]}</Badge>
      </div>
    );
  },
};
