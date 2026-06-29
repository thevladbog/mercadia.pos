import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';

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
import { getStorybookLocale } from './locale.js';

const containerCopy = {
  en: {
    sales: 'Sales',
    salesSubtitle: 'Current operational day',
    receipts: '128 receipts',
    payments: 'Payments',
    paymentsSubtitle: 'Settled successfully',
    fiscalQueue: 'Fiscal queue',
    fiscalSubtitle: 'Needs operator attention',
    pending: '3 pending',
    receiptSections: 'Receipt sections',
    lines: 'Lines',
    linesContent: 'Receipt line table placeholder',
    fiscal: 'Fiscal',
    fiscalContent: 'Fiscal document placeholder',
    paymentsContent: 'Payment state placeholder',
    openDialog: 'Open dialog',
    closeShift: 'Close shift',
    closeShiftDescription:
      'Verify cash balance, open payments, and fiscal documents before closing this shift.',
    cancel: 'Cancel',
  },
  ru: {
    sales: 'Продажи',
    salesSubtitle: 'Текущий операционный день',
    receipts: '128 чеков',
    payments: 'Платежи',
    paymentsSubtitle: 'Успешно проведены',
    fiscalQueue: 'Фискальная очередь',
    fiscalSubtitle: 'Требует внимания оператора',
    pending: '3 ожидают',
    receiptSections: 'Разделы чека',
    lines: 'Строки',
    linesContent: 'Заглушка таблицы строк чека',
    fiscal: 'Фискализация',
    fiscalContent: 'Заглушка фискального документа',
    paymentsContent: 'Заглушка состояния платежей',
    openDialog: 'Открыть диалог',
    closeShift: 'Закрыть смену',
    closeShiftDescription:
      'Проверьте остаток наличных, открытые платежи и фискальные документы перед закрытием смены.',
    cancel: 'Отмена',
  },
};

const meta = {
  title: 'Components/Containers',
  parameters: {
    layout: 'centered',
  },
} satisfies Meta;

export default meta;

type Story = StoryObj<typeof meta>;

export const Cards: Story = {
  render: (_args, context) => {
    const copy = containerCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div className="mercadia-story-grid" style={{ width: 760 }}>
        <Card>
          <CardHeading title={copy.sales} subtitle={copy.salesSubtitle} />
          <strong>{copy.receipts}</strong>
        </Card>
        <Card variant="success">
          <CardHeading title={copy.payments} subtitle={copy.paymentsSubtitle} />
          <strong>98.4%</strong>
        </Card>
        <Card variant="warning">
          <CardHeading title={copy.fiscalQueue} subtitle={copy.fiscalSubtitle} />
          <strong>{copy.pending}</strong>
        </Card>
      </div>
    );
  },
};

export const Tabs: Story = {
  render: function Render(_args, context) {
    const [activeTab, setActiveTab] = useState('lines');
    const copy = containerCopy[getStorybookLocale(context.globals.locale)];
    return (
      <div className="mercadia-story-panel" style={{ width: 640 }}>
        <PillTabs
          aria-label={copy.receiptSections}
          items={[
            {
              value: 'lines',
              label: copy.lines,
              count: 12,
              content: copy.linesContent,
            },
            {
              value: 'payments',
              label: copy.payments,
              count: 2,
              content: copy.paymentsContent,
            },
            { value: 'fiscal', label: copy.fiscal, count: 1, content: copy.fiscalContent },
          ]}
          onValueChange={setActiveTab}
          value={activeTab}
        />
      </div>
    );
  },
};

export const Dialogs: Story = {
  render: (_args, context) => {
    const copy = containerCopy[getStorybookLocale(context.globals.locale)];
    return (
      <Dialog>
        <DialogTrigger asChild>
          <Button>{copy.openDialog}</Button>
        </DialogTrigger>
        <DialogContent aria-describedby={undefined}>
          <DialogTitle>{copy.closeShift}</DialogTitle>
          <DialogBody>{copy.closeShiftDescription}</DialogBody>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="secondary">{copy.cancel}</Button>
            </DialogClose>
            <Button>{copy.closeShift}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  },
};
