import type {
  ListCashMovements200ItemsItem,
  ListCashRecounts200ItemsItem,
} from '@mercadia/api-clients-store-edge';

export function matchesMovementSearch(
  movement: ListCashMovements200ItemsItem,
  query: string,
): boolean {
  if (query.length === 0) {
    return true;
  }

  const haystack = [
    movement.id,
    movement.type,
    movement.status,
    movement.actorId,
    movement.fromContainerId,
    movement.fromContainerType,
    movement.toContainerId,
    movement.toContainerType,
    movement.reason ?? '',
    movement.approvedById ?? '',
  ]
    .join(' ')
    .toLowerCase();

  return haystack.includes(query);
}

export function matchesRecountSearch(
  recount: ListCashRecounts200ItemsItem,
  query: string,
): boolean {
  if (query.length === 0) {
    return true;
  }

  const haystack = [
    recount.id,
    recount.status,
    recount.resolutionStatus,
    recount.actorId,
    recount.containerId,
    recount.containerType,
    recount.reason ?? '',
    recount.approvedById ?? '',
    recount.resolutionNote ?? '',
    recount.resolvedById ?? '',
  ]
    .join(' ')
    .toLowerCase();

  return haystack.includes(query);
}
