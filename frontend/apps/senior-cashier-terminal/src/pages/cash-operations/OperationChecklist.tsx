import { useTranslation } from 'react-i18next';

export interface OperationChecklistProps {
  /** Already-translated by the caller — the steps differ per page and
   * don't map onto one shared, parameterized copy. */
  steps: string[];
}

/**
 * The numbered "ЧТО ДЕЛАТЬ" (what to do) list (design screens 03a/03b/03c —
 * see plan 022 Scope item 3). Purely presentational: renders whatever steps
 * the page passes in, in order.
 */
export function OperationChecklist({ steps }: OperationChecklistProps) {
  const { t } = useTranslation();

  return (
    <div className="sr-panel sr-op-checklist">
      <h2 className="sr-panel-title">{t('cash.checklist.title')}</h2>
      <ol className="sr-op-checklist-list">
        {steps.map((step, index) => (
          <li className="sr-op-checklist-item" key={index}>
            <span className="sr-op-checklist-index">{index + 1}</span>
            <span>{step}</span>
          </li>
        ))}
      </ol>
    </div>
  );
}
