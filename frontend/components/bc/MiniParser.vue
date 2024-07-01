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
import { BcLink } from '#components'
import { Target } from '~/types/links'

const props = defineProps<{ input: string|string[], links?: Record<string, string> }>()

const parsingResult = computed(() => {
  if (!Array.isArray(props.input)) {
    if (typeof props.input !== 'string') {
      return []
    }
    inputArray.lines = (props.input.includes('\r\n')) ? props.input.split('\r\n') : props.input.split('\n')
  } else {
    inputArray.lines = props.input
  }
  return parse()
})

const Escapement = '\\'
enum Tag { H1 = '#', H2 = '##', H3 = '###', Item = '-', Italic = '_', Bold = '*', Code = '`', LinkStart = '[', LinkMid = '](', LinkEnd = ')' }
const OpeningTags = [Tag.Italic, Tag.Bold, Tag.Code, Tag.LinkStart]
const FormatToStyle: Record<string, string> = { italic: 'font-style: italic;', bold: 'font-weight: bold;', code: 'font-family: monospace;' }

type Formatting = { italic: boolean, bold: boolean, code: boolean }
type Text = { text: string, format: Formatting}
type Link = { link: string, caption: Text[], target: Target}
type Block = Array<Text|Link>
type List = { list: Block[] }
type Title = { title: Block, height: number}
type Parsed = Array<Title|List|Block>

const inputArray = { lines: [] as string[], pos: 0 as number }

/** converts the input into a tree that will be very easy to browse afterwards to generate a v-DOM */
function parse () : Parsed {
  const output: Parsed = []
  for (let i = 0; i < inputArray.lines.length; i++) {
    inputArray.lines[i] = replaceEscapements(inputArray.lines[i])
  }
  inputArray.pos = 0
  while (inputArray.pos < inputArray.lines.length) {
    const line = inputArray.lines[inputArray.pos]
    if (isLineATitle(line)) {
      output.push(parseTitle({ title: [], height: 0 }, line))
      inputArray.pos++
    } else
      if (isLineAnItemInList(line)) {
        output.push(parseList({ list: [] }))
      } else {
        output.push(parseText([], line))
        inputArray.pos++
      }
  }
  return output
}

function parseTitle (output: Title, line: string) : Title {
  const { h, tag } = isLineATitle(line)!
  output.title = parseText(output.title, line.slice(tag.length))
  output.height = h
  return output
}

function parseList (output: List) : List {
  while (inputArray.pos < inputArray.lines.length) {
    if (!isLineAnItemInList(inputArray.lines[inputArray.pos])) {
      return output // end of list
    }
    output.list.push(parseText([], inputArray.lines[inputArray.pos].slice(Tag.Item.length))) // new item
    inputArray.pos++
  }
  return output
}

function parseText (output: Block, text: string, format: Formatting = { italic: false, bold: false, code: false }) : Block {
  const { pos: openingTagPos, tag: openingTag } = findTag(text, 0)
  if (openingTagPos < 0) { // no tag found
    addTextPart(output, text, format)
    return output
  }
  // opening tag found
  let closingTag: Tag
  const middle = { ...format }
  let middleIsAlink = false
  switch (openingTag) {
    case Tag.Italic : middle.italic = true; closingTag = Tag.Italic; break
    case Tag.Bold : middle.bold = true; closingTag = Tag.Bold; break
    case Tag.Code : middle.code = true; closingTag = Tag.Code; break
    case Tag.LinkStart : middleIsAlink = true; closingTag = Tag.LinkEnd; break
    default: return output
  }
  const { pos: closingTagPos } = findTag(text, openingTagPos + openingTag.length, closingTag)
  if (closingTagPos < 0) { // syntax error: either the closing tag has been forgotten or nested tags have their closure swapped
    addTextPart(output, text, format) // so we output the text without parsing it
    return output
  }
  parseLeftMiddleRightTexts(output, text, [openingTagPos, openingTagPos + openingTag.length, closingTagPos, closingTagPos + closingTag.length], format, middle, middleIsAlink)
  return output
}

function parseLeftMiddleRightTexts (output: Block, text: string, innerEnds: number[], leftRight: Formatting, middle: Formatting, middleIsAlink: boolean) : void {
  if (innerEnds[0] > 0) { addTextPart(output, text.slice(0, innerEnds[0]), leftRight) }
  if (middleIsAlink) {
    parseLink(output, text.slice(innerEnds[1], innerEnds[2]), middle)
  } else
    if (middle.code) {
      addTextPart(output, text.slice(innerEnds[1], innerEnds[2]), middle) // we do not parse anything inside
    } else {
      parseText(output, text.slice(innerEnds[1], innerEnds[2]), middle)
    }
  if (innerEnds[3] < text.length) { parseText(output, text.slice(innerEnds[3]), leftRight) }
}

function parseLink (output: Block, text: string, format: Formatting) {
  // note: param `text` is of the form  `caption of the link](urlRef`  (both ends have been removed by the calling function)
  const { pos: middleTagPos } = findTag(text, 0, Tag.LinkMid)
  const caption = parseText([], text.slice(0, middleTagPos), format) as Text[]
  const urlRef = text.slice(middleTagPos + Tag.LinkMid.length)
  const link = props.links && urlRef in props.links ? props.links[urlRef] : urlRef
  const target = (link.includes('://') || link.startsWith('www.')) ? Target.External : Target.Internal // correct 99% of the time I suppose
  output.push({ caption, link, target })
}

function isLineATitle (line: string) : { h: number, tag: Tag } | undefined {
  for (const title of [{ h: 3, tag: Tag.H3 }, { h: 2, tag: Tag.H2 }, { h: 1, tag: Tag.H1 }]) {
    if (line.startsWith(title.tag)) { return title }
  }
}

const isLineAnItemInList = (line: string) => line.startsWith(Tag.Item)

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

const addTextPart = (output: Block, text: string, format: Formatting) => output.push({ text: removeEscapements(text), format: { ...format } })

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

const removeEscapements = (raw: string) => raw.replaceAll(ESC, '')

function getLineStatus (parsed: Parsed, line: Block, pos: number) : 'skip'|'blank'|'ok' {
  if (line.length > 1 || !('text' in line[0]) || (line[0] as Text).text) {
    return 'ok'
  }
  for (const skippingReasons of ['title']) { // in the future, if new structures allow surrounding blank lines to be skipped, add them to this list
    if ((pos === 0 || pos === parsed.length - 1 || skippingReasons in parsed[pos - 1] || skippingReasons in parsed[pos + 1])) { return 'skip' }
  }
  return 'blank'
}

function RenderAll (props: {parsed: Parsed}) : any[] {
  const output = []
  for (let k = 0; k < props.parsed.length; k++) {
    const section = props.parsed[k]
    if ('title' in section) {
      output.push(h('h' + section.height, {}, renderBlock(section.title)))
    } else
      if ('list' in section) {
        const list = section.list.map(item => h('li', {}, renderBlock(item)))
        output.push(h('ul', {}, list))
      } else {
        switch (getLineStatus(props.parsed, section, k)) {
          case 'blank' : output.push(h('br', {})); break
          case 'ok' : output.push(h('div', {}, renderBlock(section))); break
        }
      }
  }
  return output
}

function renderBlock (block: Block) : any[] {
  const output = []
  for (const part of block) {
    if ('link' in part) {
      output.push(h(BcLink, { to: part.link, target: part.target, class: 'link' }, () => renderBlock(part.caption)))
    } else {
      const style = Object.entries(part.format).filter(form => form[1]).map(form => FormatToStyle[form[0]]).join(' ')
      const text = (part.format.code) ? part.text.replaceAll(' ', '\xA0') : part.text
      output.push((!style.length) ? text : h('span', { style }, text))
    }
  }
  return output
}
</script>

<template>
  <div>
    <RenderAll :parsed="parsingResult" />
  </div>
</template>

<style lang="scss">
ul {
  padding: 0;
  margin: 0;
  padding-left: 1.4em;
}
</style>
