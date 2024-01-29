import { type Currency } from './currencies'

type PriceData = {
    'symbol': string;
    'roundPrice': number;
    'truncPrice': string;
}

export type LatestState = {
    'lastProposedSlot': number;
    'currentSlot': number;
    'currentEpoch': number;
    'currentFinalizedEpoch': number;
    'finalityDelay': number;
    'syncing': boolean;
    'rates': {
        'tickerCurrency': Currency;
        'tickerCurrencySymbol': string;
        'selectedCurrency': Currency;
        'selectedCurrencySymbol': string;
        'mainCurrency': Currency;
        'mainCurrencySymbol': string;
        'mainCurrencyPrice': number;
        'mainCurrencyPriceFormatted': string;
        'mainCurrencyKFormatted': string;
        'mainCurrencyTickerPrice': number;
        'mainCurrencyTickerPriceFormatted': string;
        'mainCurrencyTickerPriceKFormatted': string;
        'elCurrency': Currency;
        'elCurrencySymbol': string;
        'elCurrencyPrice': number;
        'elCurrencyPriceFormatted': string;
        'elCurrencyKFormatted': string;
        'clCurrency': Currency;
        'clCurrencySymbol': string;
        'clCurrencyPrice': number;
        'clCurrencyPriceFormatted': string;
        'clCurrencyKFormatted': string;
        'mainCurrencyTickerPrices': Record<Currency, PriceData>
    }
}
