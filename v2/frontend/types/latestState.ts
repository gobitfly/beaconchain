import { type Currency } from './currencies'

export type LatestState = {
    'lastProposedSlot': number;
    'currentSlot': number;
    'currentEpoch': number;
    'currentFinalizedEpoch': number;
    'finalityDelay': number;
    'syncing': boolean;
    'rates': Record<Currency, number>
}
