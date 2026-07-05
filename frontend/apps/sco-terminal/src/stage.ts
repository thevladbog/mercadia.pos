/**
 * SCO sale-stage state machine, per docs/sco-terminal-implementation-design.md §2
 * ("idle -> scanning -> receipt -> payment -> done") and plans/012-sco-terminal-m1-scaffold.md
 * M1 scope. `payment` and `done` are modeled now (per the plan's "keep the stage type
 * extensible" maintenance note for M2) but are not reachable in M1 — they render as
 * disabled "coming soon" states.
 */
export type ScoStage = 'idle' | 'scanning' | 'receipt' | 'payment' | 'done';

/** Stages shown in the M1 stage indicator, in display order. */
export const STAGE_INDICATOR_ORDER: readonly ScoStage[] = [
  'scanning',
  'receipt',
  'payment',
  'done',
];

const COMING_SOON_STAGES: ReadonlySet<ScoStage> = new Set<ScoStage>(['payment', 'done']);

/** True for stages that exist in the type but are not implemented until a later milestone. */
export function isStageComingSoon(stage: ScoStage): boolean {
  return COMING_SOON_STAGES.has(stage);
}

export type ScoStageEvent =
  | { type: 'start' }
  | { type: 'reviewReceipt' }
  | { type: 'resumeScanning' }
  | { type: 'cancel' };

/** Pure transition function driving the sale-stage machine from a UI event. */
export function reduceScoStage(stage: ScoStage, event: ScoStageEvent): ScoStage {
  switch (event.type) {
    case 'start':
      return stage === 'idle' ? 'scanning' : stage;
    case 'reviewReceipt':
      return stage === 'scanning' ? 'receipt' : stage;
    case 'resumeScanning':
      return stage === 'receipt' ? 'scanning' : stage;
    case 'cancel':
      return 'idle';
    default:
      return stage;
  }
}
