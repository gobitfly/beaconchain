export type AdInsertMode = 'after' | 'before' | 'insert' | 'replace'

export type AdConfiguration = {
  banner_id?: number,
  enabled: boolean,
  for_all_users: boolean,
  html_content?: string,
  insert_mode: AdInsertMode,
  jquery_selector: string,
  key: string,
  refresh_interval: number,
}
