import type { Locator } from 'playwright-core'

type ActionCallback = () => Promise<unknown> | unknown

export type Interactions = {
  check: (selector: string, options?: {
    force?: boolean,
    strict?: boolean,
    timeout?: number,
  }) => ActionCallback,

  click: (selector: string, options?: {
    button?: 'left' | 'middle' | 'right',
    clickCount?: number,
    delay?: number,
    force?: boolean,
    modifiers?: Array<'Alt' | 'Control' | 'ControlOrMeta' | 'Meta' | 'Shift'>,
    strict?: boolean,
    timeout?: number,
  }) => ActionCallback,

  frame: (frameSelector: {
    name?: string,
    url?: ((url: URL) => boolean) | RegExp | string,
  } | string) => ActionCallback,

  frameLocator: (selector: string) => ActionCallback,

  getAttribute: (selector: string, name: string, options?: { timeout?: number }) => ActionCallback,

  getByLabel: (text: RegExp | string, options?: { exact?: boolean }) => ActionCallback,

  getByPlaceholder: (text: RegExp | string, options?: { exact?: boolean }) => ActionCallback,

  getByRole: (role: 'alert' | 'alertdialog' | 'application' | 'article' | 'banner' | 'blockquote' | 'button' | 'caption' | 'cell' | 'checkbox' | 'code' | 'columnheader' | 'combobox' | 'complementary' | 'contentinfo' | 'definition' | 'deletion' | 'dialog' | 'directory' | 'document' | 'emphasis' | 'feed' | 'figure' | 'form' | 'generic' | 'grid' | 'gridcell' | 'group' | 'heading' | 'img' | 'insertion' | 'link' | 'list' | 'listbox' | 'listitem' | 'log' | 'main' | 'marquee' | 'math' | 'menu' | 'menubar' | 'menuitem' | 'menuitemcheckbox' | 'menuitemradio' | 'meter' | 'navigation' | 'none' | 'note' | 'option' | 'paragraph' | 'presentation' | 'progressbar' | 'radio' | 'radiogroup' | 'region' | 'row' | 'rowgroup' | 'rowheader' | 'scrollbar' | 'search' | 'searchbox' | 'separator' | 'slider' | 'spinbutton' | 'status' | 'strong' | 'subscript' | 'superscript' | 'switch' | 'tab' | 'table' | 'tablist' | 'tabpanel' | 'term' | 'textbox' | 'time' | 'timer' | 'toolbar' | 'tooltip' | 'tree' | 'treegrid' | 'treeitem', options?: {
    checked?: boolean,
    disabled?: boolean,
    exact?: boolean,
    expanded?: boolean,
    includeHidden?: boolean,
    level?: number,
    name?: RegExp | string,
    pressed?: boolean,
    selected?: boolean,
  }) => ActionCallback,

  getByTestId: (testId: RegExp | string) => ActionCallback,

  getByText: (text: RegExp | string, options?: { exact?: boolean }) => ActionCallback,

  getByTitle: (text: RegExp | string, options?: { exact?: boolean }) => ActionCallback,
  goBack: (options?: {
    timeout?: number,
    waitUntil?: 'commit' | 'domcontentloaded' | 'load' | 'networkidle',
  }) => ActionCallback,
  goForward: (options?: {
    timeout?: number,
    waitUntil?: 'commit' | 'domcontentloaded' | 'load' | 'networkidle',
  }) => ActionCallback,
  type: (selector: string, value: string, options?: {
    force?: boolean,
    timeout?: number,
  }) => ActionCallback,

}

export type LocatorAssertions = {
  goto: (url: string, options?: {
    referer?: string,
    timeout?: number,
    waitUntil?: 'commit' | 'domcontentloaded' | 'load' | 'networkidle',
  }) => ActionCallback,
  locator: (selector: string, options?: {
    has?: Locator,
    hasNot?: Locator,
    hasNotText?: RegExp | string,
    hasText?: RegExp | string,
  }) => ActionCallback,
  reload: (options?: {
    timeout?: number,
    waitUntil?: 'commit' | 'domcontentloaded' | 'load' | 'networkidle',
  }) => ActionCallback,
  shouldBeChecked: (options?: { timeout?: number }) => ActionCallback,
  shouldBeDisabled: (options?: { timeout?: number }) => ActionCallback,
  shouldBeEmpty: (options?: { timeout?: number }) => ActionCallback,
  shouldBeHidden: (options?: { timeout?: number }) => ActionCallback,
  shouldBeVisible: (options?: { timeout?: number }) => ActionCallback,
  shouldContainText: (expected: ReadonlyArray<RegExp | string> | RegExp | string, options?: {
    ignoreCase?: boolean,
    timeout?: number,
    useInnerText?: boolean,
  }) => ActionCallback,
  shouldHaveAttribute: (name: string, value: RegExp | string, options?: {
    ignoreCase?: boolean,
    timeout?: number,
  }) => ActionCallback,
  shouldHaveValue: (value: RegExp | string, options?: { timeout?: number }) => ActionCallback,
  shouldHaveValues: (values: ReadonlyArray<RegExp | string>, options?: { timeout?: number }) => ActionCallback,
  shouldToBeEnabled: (options?: { timeout?: number }) => ActionCallback,
  shouldToHaveText: (name: ReadonlyArray<RegExp | string> | RegExp | string, options?: {
    ignoreCase?: boolean,
    timeout?: number,
    useInnerText?: boolean,
  }) => ActionCallback,
  uncheck: (selector: string, options?: {
    force?: boolean,
    strict?: boolean,
    timeout?: number,
  }) => ActionCallback,
}

export type PageAssertions = {
  shouldHaveTitle: (titleOrRegExp: RegExp | string, options?: { timeout?: number }) => ActionCallback,
  shouldHaveURL: (urlOrRegExp: RegExp | string, options?: { timeout?: number }) => ActionCallback,
}

export type AssertionsNot = {
  shouldNotBeVisible: () => ActionCallback,
  shouldNotExist: () => ActionCallback,
}

type FindByLabelText = (text: string) => AssertionsNot & Interactions

type Role = 'button' | 'link' | 'option' | 'tab'

type FindByRoleOptions = {
  name: string,
}

type FindByRole = (role: Role, options: FindByRoleOptions) => AssertionsNot & Interactions

type FindByTestId = (testId: string) => AssertionsNot & Interactions

type FindByTextOptions = {
  withinTestId?: string,
}

type FindByText = (text: string, options?: FindByTextOptions) => AssertionsNot

type FindAllByText = (text: string, options?: FindByTextOptions) => AssertionsNot

type QueryByText = (text: string, options?: FindByTextOptions) => AssertionsNot

export type GoToOptions = {
  device?: 'desktop' | 'mobile',
}

type GoTo = (path: string, options?: GoToOptions) => () => void

type Body = Record<number | string, unknown>

type GetBody = ({ searchParams }: { searchParams: URLSearchParams }) => Body

export type MockEndpoint = (endpoint: string, options: {
  body: Body | GetBody,
  httpVerb: 'delete' | 'get' | 'patch' | 'post',
  status: number,
}) => void

type PreconditionOptions = {
  localStorage: typeof window.localStorage,
  mockEndpoint: MockEndpoint,
}

export type Precondition = (options: PreconditionOptions) => void

export type Prepare = (precondition: Precondition) => () => void

export type Driver = {
  findAllByText: FindAllByText,
  findByLabelText: FindByLabelText,
  findByRole: FindByRole,
  findByTestId: FindByTestId,
  findByText: FindByText,
  goTo: GoTo,
  prepare: Prepare,
  queryByText: QueryByText,
}
