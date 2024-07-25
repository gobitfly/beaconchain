<script setup lang="ts">
/** Component outputting HTML by interpreting simple markdown-like tags found in the input.
 *
 *  Usage:
 *
 *    Simplest case:
 *      <BcMiniParser :input="stringOrArrayOfStrings" />
 *    You can pass an object containing variables whose values will be inserted:
 *      <BcMiniParser :input="stringOrArrayOfStrings" :variables="{darkLink: 'www.choco.com', easterDuration: 50}" />
 *    (see the example just below to see how to use the variables)
 *
 *  Example of input:
 *
 *    # About chocolate
 *    We will eat `chocolate` in 2 cases:
 *    - If your *validator* is _online_
 *    - During the $easterDuration days of Easter\*.
 *    It can be [very dark]($darkLink) or with milk, we enjoy both.
 *    \*note that Christmas is also _a good moment_ to do so
 *
 *  Language:
 *
 *    The input can be either a string (that can hold several lines of text) or an array of strings (each element is a single line of text).
 *    For each line:
 *      # or ## or ### at the beginning will show the line as a title (respectively h1, h2, h3)
 *      - at the beginning will show the line as an item in a list.
 *      parts surrounded with _ will be surrounded with <i> tags
 *      parts surrounded with * will be surrounded with <b> tags
 *      parts surrounded with ` will be shown with a type-writter font and not parsed (formatting tags between ` and ` are ineffective)
 *      a link can be created by writing [a caption](and-a-url). The url can written directly or come from an entry in the object in prop :variables.
 *    When numbers are inserted from prop :variables, they are formatted according to the local settings of the user.
 *    Mixes are possible: italic inside bold or bold inside italic, code in italic or bold (if you surround ` and ` with the tags)...
 *
 *    If you need to display a character that is a tag, escape it with `\` and the parser will not interpret it.
 *    As Javascript itself uses `\` as an escaping mark, you will need to type `\\` and `\\\\` to express respectively `\` and `\\` when you
 *    hard-code the input in JS.
 *
 *  About <i> and <b>:
 *
 *    Depending on the stylesheet of your website, those tags might not display your text in italic and bold.
 *    If you want to force them to display italic and bold text against the style preferences of your website, you can define a class like so:
 *      .my-own-bi {
 *        :deep(i) { font-style: italic }
 *        :deep(b) { font-weight: bold }
 *      }
 *    and assign it to the parser:
 *      <BcMiniParser ... class="my-own-bi" />
*/

import { type VNode } from 'vue'
import { BcLink } from '#components'
import { Target } from '~/types/links'

const Escapement = '\\'
enum Tag { H1 = '#', H2 = '##', H3 = '###', Item = '-', Italic = '_', Bold = '*', Code = '`', LinkStart = '[', LinkMid = '](', LinkEnd = ')', Variable = '$' }
const OpeningTags = [Tag.Italic, Tag.Bold, Tag.Code, Tag.LinkStart]
const ClosingTags: Record<string, Tag> = { [Tag.Italic]: Tag.Italic, [Tag.Bold]: Tag.Bold, [Tag.Code]: Tag.Code, [Tag.LinkStart]: Tag.LinkEnd }

const props = defineProps<{ input: string|string[], variables?: Record<string, string> }>()

type VDOMnodes = Array<VNode|string>
enum LineType { Useless, Blank, Title, List, Div }
const ESC = '\u001B'
const inputArray = { lines: [] as string[], pos: 0 as number }
const variableNamesSortedByLength = !props.variables ? [] : Object.keys(props.variables).sort((a, b) => b.length - a.length)

function parse (props: {input: string|string[]}) : VDOMnodes {
  if (!Array.isArray(props.input)) {
    if (typeof props.input !== 'string') { return [] }
    inputArray.lines = (props.input.includes('\r\n')) ? props.input.split('\r\n') : props.input.split('\n')
  } else {
    inputArray.lines = props.input
  }
  inputArray.lines = inputArray.lines.map(line => removeLeadingSpacesAndReplaceEscapements(line))
  inputArray.pos = 0
  const output: VDOMnodes = []
  while (inputArray.pos < inputArray.lines.length) {
    switch (getLineType()) {
      case LineType.Title : output.push(parseTitle()); break
      case LineType.List : output.push(parseList()); break
      case LineType.Blank :
      case LineType.Div : output.push(parseTextLine()); break
      default: inputArray.pos++
    }
  }
  return output
}

function parseTitle () : VNode {
  const { height, tag } = getTitleType(inputArray.pos)!
  return h('h' + height, {}, parseText(inputArray.lines[inputArray.pos++].slice(tag.length)))
}

function parseList () : VNode {
  const items : VDOMnodes = []
  while (inputArray.pos < inputArray.lines.length && getLineType() === LineType.List) {
    items.push(h('li', {}, parseText(inputArray.lines[inputArray.pos].slice(Tag.Item.length)))) // new item
    inputArray.pos++
  }
  return h('ul', {}, items)
}

function parseTextLine () : VNode|string {
  let output: VNode|string = ''
  switch (getLineType()) {
    case LineType.Div : output = h('div', {}, parseText(inputArray.lines[inputArray.pos])); break
    case LineType.Blank : output = h('br', {}); break
  }
  inputArray.pos++
  return output
}

function parseText (text: string) : VDOMnodes {
  const output: VDOMnodes = []
  do {
    const { pos: openingTagPos, tag: openingTag } = findTag(text, 0)
    const { pos: closingTagPos } = findTag(text, openingTagPos + openingTag.length, ClosingTags[openingTag])
    if (openingTagPos < 0 || closingTagPos < 0) { // First case: no tag, we can copy the raw line. Second case: syntax error (either the closing tag has been forgotten or nested tags have their closure swapped)
      output.push(cleanText(text)) // in both cases we output the text without parsing it
      break
    }
    if (openingTagPos > 0) {
      output.push(cleanText(text.slice(0, openingTagPos)))
    }
    const middle = text.slice(openingTagPos + openingTag.length, closingTagPos)
    switch (openingTag) {
      case Tag.Italic : output.push(h('i', {}, parseText(middle))); break
      case Tag.Bold : output.push(h('b', {}, parseText(middle))); break
      case Tag.Code : output.push(h('span', { style: 'font-family: monospace;' }, cleanText(middle, true))); break
      case Tag.LinkStart : output.push(parseLink(middle)); break
    }
    text = text.slice(closingTagPos + ClosingTags[openingTag].length)
  } while (text)
  return output
}

function parseLink (text: string) : VNode|string {
  // note: param `text` is of the form  `caption of the link](urlRef`  (both ends have been removed by the calling function)
  const { pos: middleTagPos } = findTag(text, 0, Tag.LinkMid)
  if (middleTagPos <= 0) { // syntax error (the link has no caption) : we output the text without parsing it
    return cleanText(text)
  }
  const urlRef = text.slice(middleTagPos + Tag.LinkMid.length)
  const to = urlRef.startsWith(Tag.Variable) ? getVariableValue(urlRef) : urlRef
  const target = (to.includes('://') || to.startsWith('www.')) ? Target.External : Target.Internal // correct 99% of the time I suppose
  return h(BcLink, { to, target, class: 'link' }, () => parseText(text.slice(0, middleTagPos)))
}

/** @param wanted If given, finds the first valid occurence of this tag (this mode is used for closing tags). If omitted, finds the first opening tag that the parser recognises.
 *  @returns -1 if not found */
function findTag (text: string, start: number, wanted?: Tag) : { pos: number, tag: Tag } {
  if (wanted !== undefined) {
    let pos = start - wanted.length
    do {
      pos = text.indexOf(wanted, pos + wanted.length)
    } while (pos >= 1 && (text[pos - 1] === ESC || (text.slice(start, pos).split(Tag.Code).length - text.slice(start, pos).split(ESC + Tag.Code).length) % 2)) // are ignored: escaped tags and tags between ` and `
    return { pos, tag: wanted }
  }
  let closest = { pos: -1, tag: Tag.Italic }
  for (const tag of OpeningTags) {
    const found = findTag(text, start, tag)
    if (found.pos >= 0 && (found.pos <= closest.pos || closest.pos < 0)) {
      closest = found
    }
  }
  return closest
}

/** replaces all `\` with `\u001B` and all `\\` with `\`  and removes spaces at the beginning */
function removeLeadingSpacesAndReplaceEscapements (input: string) : string {
  const DoubleEsc = Escapement + Escapement
  let posIn = 0
  let output = ''
  while (posIn < input.length) {
    const double = input.indexOf(DoubleEsc, posIn)
    const single = input.indexOf(Escapement, posIn)
    if (double >= 0 && (double <= single || single < 0)) {
      output += input.slice(posIn, double) + Escapement
      posIn = double + DoubleEsc.length
    } else
      if (single >= 0) {
        output += input.slice(posIn, single) + ESC
        posIn = single + Escapement.length
      } else {
        break
      }
  }
  return (output + input.slice(posIn)).trimStart()
}

function cleanText (raw: string, forceSpaces = false) : string {
  if (forceSpaces) {
    raw = raw.replaceAll(' ', '\xA0')
  }
  return raw.replaceAll(ESC, '')
}

function getVariableValue (name: string) : any {
  if (!props.variables) { return name }
  const key = name.startsWith(Tag.Variable) ? name.slice(Tag.Variable.length) : name
  for (const v of variableNamesSortedByLength) {
    if (key.startsWith(v)) { return props.variables[v] }
  }
  return name
}

function getTitleType (pos: number) : { height: number, tag: Tag } | undefined {
  for (const title of [{ height: 3, tag: Tag.H3 }, { height: 2, tag: Tag.H2 }, { height: 1, tag: Tag.H1 }]) {
    if (inputArray.lines[pos].startsWith(title.tag)) { return title }
  }
}

function getLineType () : LineType {
  const pos = inputArray.pos
  if (inputArray.lines[pos].length > 0) {
    if (getTitleType(pos)) { return LineType.Title }
    if (inputArray.lines[pos].startsWith(Tag.Item)) { return LineType.List }
    return LineType.Div
  }
  if (pos === 0 || pos === inputArray.lines.length - 1 || getTitleType(pos - 1) || getTitleType(pos + 1)) { return LineType.Useless }
  return LineType.Blank
}
</script>

<template>
  <div>
    <parse :input="props.input" />
  </div>
</template>

<style lang="scss">
ul {
  padding: 0;
  margin: 0;
  padding-left: 1.4em;
}
</style>
