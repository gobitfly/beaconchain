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
 *    `-` at the beginning transforms the line into a list item.
 *      words surrounded with _ will be shown in italic
 *      words surrounded with * will be shown in bold
 *      words surrounded with ` will be shown with a type-writter font and not parsed (formatting tags inside ` and ` are ineffective)
 *      a link can be created by writing [a caption](and-a-url). The url can written directly or tell the name of a member of the object in props links.
 *    Mixes are possible: italic inside bold or bold inside italic.
 *
 *    If you need to display a character that is a tag, escape it with `\` and the parser will not interpret it.
 *    As Javascript itself uses `\` as an escaping mark, you will need to type `\\` and `\\\\` to express respectively `\` and `\\` when you
 *    hard-code the input in JS.
 *
*/
import { Target } from '~/types/links'

const props = defineProps<{ input: string|string[], links?: Record<string, string> }>()

const parsed = computed(() => parse(props.input))

enum Tag { Italic = '_', Bold = '*', Code = '`', Item = '-', LinkStart = '[', LinkMid = '](', LinkEnd = ')' }
const OpeningTags = [Tag.Italic, Tag.Bold, Tag.Code, Tag.LinkStart]
const Escapement = '\\'

type Formatting = { italic: boolean, bold: boolean, code: boolean }
type Text = { content: string, format: Formatting}
type Link = { caption: Text[], url: string, target: Target}
type LinePart = Text | Link
type Line = LinePart[]
type List = Line[]
type Parsed = Array<Line|List>

/** converts the input into a structure that will be very easy to output in `<template>` */
function parse (raw : string|string[]) : Parsed {
  const output: Parsed = []
  if (!Array.isArray(raw)) {
    if (typeof raw !== 'string') {
      return []
    }
    raw = (raw.includes('\r\n')) ? raw.split('\r\n') : raw.split('\n')
  }
  if (!raw.length || (raw.length === 1 && !raw[0])) {
    return []
  }
  const cleaned: string[] = []
  for (let i = 0; i < raw.length; i++) {
    cleaned.push(replaceEscapements(raw[i]))
  }
  let i = 0
  while (i < cleaned.length) {
    if (isLineAnItemInList(cleaned[i])) {
      const list: List = []
      output.push(list)
      i = parseList(list, cleaned, i)
    } else {
      const line: Line = []
      output.push(line)
      parseText(line, cleaned[i])
      i++
    }
  }
  return output
}

function parseList (output: List, raw : string[], start : number) : number {
  let i = start
  while (i < raw.length) {
    if (!isLineAnItemInList(raw[i])) {
      return i // end of list, we return the index of this first text after the list
    }
    const item: Line = [] // new item in the list
    output.push(item)
    parseText(item, raw[i].slice(Tag.Item.length))
    i++
  }
  return i
}

function parseText (output: Line, text: string, format: Formatting = { italic: false, bold: false, code: false }) : void {
  const { pos: openingTagPos, tag: openingTag } = findFirstTag(text, 0)
  if (openingTagPos < 0) { // no tag found
    addTextPart(output, text, format)
    return
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
    default: return
  }
  const { pos: closingTagPos } = findFirstTag(text, openingTagPos + openingTag.length, closingTag)
  if (closingTagPos < 0) { // syntax error: either the closing tag has been forgotten or nested tags have their closure swapped
    addTextPart(output, text, format) // so we output the text without parsing it
    return
  }
  parseLeftMiddleRightTexts(output, text, [openingTagPos, openingTagPos + openingTag.length, closingTagPos, closingTagPos + closingTag.length], format, middle, middleIsAlink)
}

function parseLeftMiddleRightTexts (output: Line, text: string, innerEnds: number[], leftRight: Formatting, middle: Formatting, middleIsAlink: boolean) {
  if (innerEnds[0] > 0) {
    addTextPart(output, text.slice(0, innerEnds[0]), leftRight)
  }
  if (middleIsAlink) {
    parseLink(output, text.slice(innerEnds[1], innerEnds[2]), middle)
  } else
    if (middle.code) {
      addTextPart(output, text.slice(innerEnds[1], innerEnds[2]), middle) // we do not parse anything inside
    } else {
      parseText(output, text.slice(innerEnds[1], innerEnds[2]), middle)
    }
  if (innerEnds[3] < text.length) {
    parseText(output, text.slice(innerEnds[3]), leftRight)
  }
}

function parseLink (output: Line, text: string, format: Formatting) {
  // note: param `text` is of the form  `caption of the link](urlRef`  (both ends have been removed by the calling function)
  const { pos: middleTagPos } = findFirstTag(text, 0, Tag.LinkMid)
  const caption: Text[] = []
  parseText(caption, text.slice(0, middleTagPos), format)
  const urlRef = text.slice(middleTagPos + Tag.LinkMid.length)
  const url = props.links && urlRef in props.links ? props.links[urlRef] : urlRef
  const target = (url.includes('://') || url.startsWith('www.')) ? Target.External : Target.Internal // correct 99% of the time I suppose
  output.push({ caption, url, target } as Link)
}

function isLineAnItemInList (part: string) : boolean {
  return part.slice(0, Tag.Item.length) === Tag.Item // note that if the list tag is escaped, this test returns false as expected
}

const ESC = '\u001B'

/** @param wanted If given, finds the first occurence of this tag (this mode is used for closing tags). If omitted, finds the first opening tag that the parser recognises.
 *  @returns -1 if not found
 * */
function findFirstTag (text: string, start: number, wanted?: Tag) : { pos: number, tag: Tag } {
  if (wanted !== undefined) {
    let pos = start - wanted.length
    do {
      pos = text.indexOf(wanted, pos + wanted.length)
    } while (pos >= 1 && text[pos - 1] === ESC) // escaped tags are ignored
    return { pos, tag: wanted }
  }
  let closest = { pos: -1, tag: Tag.Italic }
  for (const tag of OpeningTags) {
    const found = findFirstTag(text, start, tag)
    if (found.pos >= 0 && (found.pos < closest.pos || closest.pos < 0)) {
      closest = found
    }
  }
  return closest
}

function addTextPart (output: Line, text: string, format: Formatting) :void {
  output.push({ content: removeEscapements(text), format: { ...format } })
}

/** replace all `\` with `\u001B` and all `\\` with `\`  */
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

function removeEscapements (raw: string) : string {
  return raw.replaceAll(ESC, '')
}
</script>

<template>
  <div>
    <div v-for="(element,i) of parsed" :key="i">
      <ul v-if="Array.isArray(element[0])">
        <li v-for="(item,j) of (element as List)" :key="j">
          <span v-for="(part,k) of item" :key="k">
            <span v-if="!('url' in part)" :class="part.format">{{ part.content }}</span>
            <BcLink v-else :to="part.url" class="link" :target="part.target">
              <span v-for="(text,l) of part.caption" :key="l" :class="text.format">{{ text.content }}</span>
            </BcLink>
          </span>
        </li>
      </ul>
      <br v-else-if="element.length <= 1 && 'content' in element[0] && !(element[0] as Text).content">
      <span v-for="(part,k) of (element as Line)" v-else :key="k">
        <span v-if="!('url' in part)" :class="part.format">{{ part.content }}</span>
        <BcLink v-else :to="part.url" class="link" :target="part.target">
          <span v-for="(text,l) of part.caption" :key="l" :class="text.format">{{ text.content }}</span>
        </BcLink>
      </span>
    </div>
  </div>
</template>

<style scoped lang="scss">
ul {
  padding: 0;
  margin: 0;
  padding-left: 1.4em;
}
.italic {
  font-style: italic;
}
.bold {
  font-weight: bold;
}
.code {
  font-family: monospace;
}
</style>
