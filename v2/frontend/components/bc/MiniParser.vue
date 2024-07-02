<script setup lang="ts">
/** Component outputting HTML by interpreting simple tags found in the input.
 *
 *  Usage:
 *
 *    Simplest case:
 *      <BcMiniParser :input="stringOrArrayOfStrings" />
 *    You can pass an object containing links:
 *      <BcMiniParser :input="stringOrArrayOfStrings" :links="{darklink: 'www.choco.com', milkyurl: 'index.html'}" />
 *    (see the example just below to understand the purpose of prop links)
 *
 *  Example of input:
 *
 *    # About chocolate
 *    We will eat `chocolate` in 2 cases:
 *    - If your *validator* is online
 *    - During *the _nice_ days* of Easter\*.
 *    It can be [very dark](darklink) or [with _milk_](milkyurl), we enjoy both.
 *    \*note that Christmas is also a good moment to do so
 *
 *  Language:
 *
 *    The input can be either a string (that can hold several lines of text) or an array of strings (each element is a single line of text).
 *    For each line:
 *      # or ## or ### at the beginning will show the line as a title (respectively h1, h2, h3)
 *      - at the beginning will show the line as an item in a list.
 *      words surrounded with _ will be shown in italic
 *      words surrounded with * will be shown in bold
 *      words surrounded with ` will be shown with a type-writter font and not parsed (formatting tags inside ` and ` are ineffective)
 *      a link can be created by writing [a caption](and-a-url). The url can written directly or tell the name of a member of the object in props links.
 *    Mixes are possible: italic inside bold or bold inside italic, code in italic or bold (if you surround ` and ` with the tags)...
 *
 *    If you need to display a character that is a tag, escape it with `\` and the parser will not interpret it.
 *    As Javascript itself uses `\` as an escaping mark, you will need to type `\\` and `\\\\` to express respectively `\` and `\\` when you
 *    hard-code the input in JS.
 *
*/

import { type VNode } from 'vue'
import { BcLink } from '#components'
import { Target } from '~/types/links'

const props = defineProps<{ input: string|string[], links?: Record<string, string> }>()

const Escapement = '\\'
enum Tag { H1 = '#', H2 = '##', H3 = '###', Item = '-', Italic = '_', Bold = '*', Code = '`', LinkStart = '[', LinkMid = '](', LinkEnd = ')' }
const OpeningTags = [Tag.Italic, Tag.Bold, Tag.Code, Tag.LinkStart]
const ClosingTags: Record<string, Tag> = { [Tag.Italic]: Tag.Italic, [Tag.Bold]: Tag.Bold, [Tag.Code]: Tag.Code, [Tag.LinkStart]: Tag.LinkEnd }

type VDOMnodes = Array<VNode|string>

const inputArray = { lines: [] as string[], pos: 0 as number }

/** converts the input into a tree that will be very easy to browse afterwards to generate a v-DOM */
function parse (props: {input: string|string[]}) : VDOMnodes {
  if (!Array.isArray(props.input)) {
    if (typeof props.input !== 'string') { return [] }
    inputArray.lines = (props.input.includes('\r\n')) ? props.input.split('\r\n') : props.input.split('\n')
  } else {
    inputArray.lines = props.input
  }
  inputArray.lines = inputArray.lines.map(line => replaceEscapements(line))
  inputArray.pos = 0
  const output: VDOMnodes = []
  while (inputArray.pos < inputArray.lines.length) {
    if (isLineATitle(inputArray.lines[inputArray.pos])) {
      output.push(parseTitle())
    } else
      if (isLineAnItemInList(inputArray.lines[inputArray.pos])) {
        output.push(parseList())
      } else
        if (lineType() !== 'none') {
          output.push(parseLine(lineType()))
        } else {
          inputArray.pos++
        }
  }
  return output
}

function parseTitle () : VNode {
  const { height, tag } = isLineATitle(inputArray.lines[inputArray.pos])!
  return h('h' + height, {}, parseText(inputArray.lines[inputArray.pos++].slice(tag.length)))
}

function parseList () : VNode {
  const items : VDOMnodes = []
  while (inputArray.pos < inputArray.lines.length && isLineAnItemInList(inputArray.lines[inputArray.pos])) {
    items.push(h('li', {}, parseText(inputArray.lines[inputArray.pos].slice(Tag.Item.length)))) // new item
    inputArray.pos++
  }
  return h('ul', {}, items)
}

function parseLine (lineType: 'full'|'blank'|'none') : VNode {
  const pos = inputArray.pos++
  switch (lineType) {
    case 'full' : return h('div', {}, parseText(inputArray.lines[pos]))
    default : return h('br', {})
  }
}

function parseText (text: string) : VDOMnodes {
  const { pos: openingTagPos, tag: openingTag } = findTag(text, 0)
  const { pos: closingTagPos } = findTag(text, openingTagPos + openingTag.length, ClosingTags[openingTag])
  if (openingTagPos < 0 || closingTagPos < 0) { // First case: no tag, we can copy the raw line. Second case: syntax error (either the closing tag has been forgotten or nested tags have their closure swapped)
    return [cleanText(text)] // so we output the text without parsing it
  }
  // opening tag found.
  const threeParts : VDOMnodes = []
  if (openingTagPos > 0) {
    threeParts.push(cleanText(text.slice(0, openingTagPos)))
  }
  const middle = text.slice(openingTagPos + openingTag.length, closingTagPos)
  switch (openingTag) {
    case Tag.Italic : threeParts.push(h('i', {}, parseText(middle))); break
    case Tag.Bold : threeParts.push(h('b', {}, parseText(middle))); break
    case Tag.Code : threeParts.push(h('span', { style: 'font-family: monospace;' }, cleanText(middle, true))); break
    case Tag.LinkStart : threeParts.push(parseLink(middle)); break
  }
  if (closingTagPos + ClosingTags[openingTag].length < text.length) {
    threeParts.push(...parseText(text.slice(closingTagPos + ClosingTags[openingTag].length)))
  }
  return threeParts
}

function parseLink (text: string) : VNode|string {
  // note: param `text` is of the form  `caption of the link](urlRef`  (both ends have been removed by the calling function)
  const { pos: middleTagPos } = findTag(text, 0, Tag.LinkMid)
  if (middleTagPos <= 0) {
    return cleanText(text)
  }
  const urlRef = text.slice(middleTagPos + Tag.LinkMid.length)
  const to = props.links && urlRef in props.links ? props.links[urlRef] : urlRef
  const target = (to.includes('://') || to.startsWith('www.')) ? Target.External : Target.Internal // correct 99% of the time I suppose
  return h(BcLink, { to, target, class: 'link' }, () => parseText(text.slice(0, middleTagPos)))
}

const ESC = '\u001B'

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
    if (found.pos >= 0 && (found.pos < closest.pos || closest.pos < 0)) {
      closest = found
    }
  }
  return closest
}

/** replaces all `\` with `\u001B` and all `\\` with `\`  */
function replaceEscapements (input: string) : string {
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
  return (output + input.slice(posIn))
}

function cleanText (raw: string, forceSpaces = false) {
  if (forceSpaces) {
    raw = raw.replaceAll(' ', '\xA0')
  }
  return raw.replaceAll(ESC, '')
}

function isLineATitle (line: string) : { height: number, tag: Tag } | undefined {
  for (const title of [{ height: 3, tag: Tag.H3 }, { height: 2, tag: Tag.H2 }, { height: 1, tag: Tag.H1 }]) {
    if (line.startsWith(title.tag)) { return title }
  }
}

const isLineAnItemInList = (line: string) => line.startsWith(Tag.Item)

function lineType () : 'full'|'blank'|'none' {
  const pos = inputArray.pos
  if (inputArray.lines[pos].length > 0) { return 'full' }
  if (pos === 0 || pos === inputArray.lines.length - 1 || isLineATitle(inputArray.lines[pos - 1]) || isLineATitle(inputArray.lines[pos + 1])) { return 'none' }
  return 'blank'
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
