import type { Meta, StoryObj } from '@storybook/react-vite';

import { Button } from '../index.js';
import { getStorybookLocale } from './locale.js';

const buttonCopy = {
  en: {
    confirm: 'Confirm payment',
    primary: 'Primary',
    secondary: 'Secondary',
    ghost: 'Ghost',
    link: 'Link',
    small: 'Small',
    medium: 'Medium',
    large: 'Large',
    add: 'Add',
  },
  ru: {
    confirm: 'Подтвердить оплату',
    primary: 'Основная',
    secondary: 'Вторичная',
    ghost: 'Прозрачная',
    link: 'Ссылка',
    small: 'Маленькая',
    medium: 'Средняя',
    large: 'Большая',
    add: 'Добавить',
  },
};

const meta = {
  title: 'Components/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'ghost', 'link'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg', 'icon'],
    },
  },
  args: {
    children: 'Confirm payment',
    variant: 'primary',
    size: 'md',
  },
} satisfies Meta<typeof Button>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args, context) => {
    const copy = buttonCopy[getStorybookLocale(context.globals.locale)];
    return <Button {...args}>{copy.confirm}</Button>;
  },
};

export const Variants: Story = {
  render: (_args, context) => {
    const copy = buttonCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div className="mercadia-story-row">
        <Button variant="primary">{copy.primary}</Button>
        <Button variant="secondary">{copy.secondary}</Button>
        <Button variant="ghost">{copy.ghost}</Button>
        <Button variant="link">{copy.link}</Button>
      </div>
    );
  },
};

export const Sizes: Story = {
  render: (_args, context) => {
    const copy = buttonCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div className="mercadia-story-row">
        <Button size="sm">{copy.small}</Button>
        <Button size="md">{copy.medium}</Button>
        <Button size="lg">{copy.large}</Button>
        <Button aria-label={copy.add} size="icon">
          +
        </Button>
      </div>
    );
  },
};
