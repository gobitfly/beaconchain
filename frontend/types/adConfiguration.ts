export type AdInsertMode = 'after' | 'before' | 'replace' | 'insert'

export type AdConfiguration = {
  'key': string,
  'jquery_selector': string,
  'insert_mode': AdInsertMode,
  'refresh_interval': number,
  'for_all_users': boolean,
  'banner_id'?: number,
  'html_content'?: string,
  'enabled': boolean
}
