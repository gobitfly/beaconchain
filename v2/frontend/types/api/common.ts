/* Do not change, this code is generated from Golang structs */


export interface Paging {
    prev_cursor: string;
    next_cursor: string;
}
export interface ApiResponse {
    paging?: Paging;
    data: any;
}
export interface ApiErrorResponse {
    error: string;
}
export interface Address {
    hash: string;
    ens?: string;
}
export interface LuckItem {
    percent: number;
    expected: number;
    average: number;
}
export interface Luck {
    proposal: LuckItem;
    sync: LuckItem;
}

export interface StatusCount {
    success: number;
    failed: number;
}
export interface ClElValue {
    el: string;
    cl: string;
}
export interface ClElValueFloat {
    el: number;
    cl: number;
}
export interface PeriodicClElValues {
    total: ClElValue;
    day: ClElValue;
    week: ClElValue;
    month: ClElValue;
    year: ClElValue;
}
export interface PeriodicClElValuesFloat {
    total: ClElValueFloat;
    day: ClElValueFloat;
    week: ClElValueFloat;
    month: ClElValueFloat;
    year: ClElValueFloat;
}
export interface HighchartsDataPoint {
    x: number;
    y: number;
}
export interface HighchartsSeries {
    name: string;
    data: HighchartsDataPoint[];
}

export interface SearchResult {
    type: string;
    chain_id: number;
    hash_value?: string;
    num_value?: number;
    str_value?: string;
}
export interface SearchResponse {
    data: SearchResult[];
}
