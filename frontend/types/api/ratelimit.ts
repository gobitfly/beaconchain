// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ApiDataResponse } from './common'

//////////
// source: ratelimit.go

export interface ApiWeightItem {
  Bucket: string;
  Endpoint: string;
  Method: string;
  Weight: number /* int */;
}
export type InternalGetRatelimitWeightsResponse = ApiDataResponse<ApiWeightItem[]>;
