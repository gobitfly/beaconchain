**Summary**

- Using the search bar in your code
- Changing the behavior and look of the search bar thanks to _searchbar.ts_
- Usage of _MiddleEllipsis.vue_

# Using the search bar in your code

In your `<script setup>`, write
```TS
const mySearchbar = ref<SearchBar>()
```

In your template, write
```HTML
<BcSearchbarMain
  ref="mySearchbar"
  :bar-style="SearchbarStyle.<what you want to see>"
  :bar-purpose="SearchbarPurpose.<what you want the bar to do>"
  :pick-by-default="<function that picks a result on behalf of the user when they press Enter or the search-button>"
  @go="<your function doing something when a result is selected>"
/>
```

There are more props that you can give to configure the search bar:
```TS
:only-networks="[<list of chain IDs that the bar is authorized to search over>]" // Without this props, the bar searches over all networks.
:keep-dropdown-open="true" // When the user selects a result, the drop-down does not close.
```

The list of possible values for `:bar-style` is in enum `SearchbarStyle` in file _searchbar.ts_.
The list of possible values for `:bar-purpose` is in enum `SearchbarPurpose` in file _searchbar.ts_.
You can write your own function for `:pick-by-default`, or give the example function written in _searchbar.ts_ if it suits your needs:
```TS
:pick-by-default="pickHighestPriorityAmongBestMatchings"
```

The handler that you give to props `@go` receives one parameter:
```TS
function myHandler (result : ResultSuggestion)
```
To get a description of the information carried in the parameter, look at the comments around the declaration of type `ResultSuggestion`.

The search-bar offers methods that you can call for further tailoring:
```TS
mySearchbar.value!.hideResult(whichOne : ResultSuggestion) // Removes a result from the drop-down. The result is one of those that you obtained in your `@go` handler.
mySearchbar.value!.closeDropdown() // Useful when you gave `:keep-dropdown-open="true"` in the props but you still need to close the drop-down in certain cases.
mySearchbar.value!.empty() // By itself, the search-bar never empties its input field nor its drop-down. You can still clear the search-bar with this method if you want the user to retype from scratch.
```

# Changing the behavior and look of the search bar thanks to _searchbar.ts_

_searchbar.ts_ has been designed as a "configuration file" for the search bar. For example, if the protocol to communicate with the API changes, it might
not be necessary to modify the code of the search-bar (its behavior and look are hard-coded as little as possible).

**If the API returns a new type of result:**
  1. Add this type into the `ResultType` enum of _searchbar.ts_.
  2. Tell the bar how to read/display it by adding a new entry in the `TypeInfo` record.

**If the API gets the ability to return a new field in some or all elements of its response array:**
  1. Write the name of the new field in `SingleAPIresult`.
  2. Create a reference to it in `Indirect`.
  3. In `TypeInfo`, tell the bar when/where this field must be read (by giving its `Indirect` reference).
  4. Add a case for the reference in function `wasOutputDataGivenByTheAPI()`
  5. Add a case for the reference in function `realizeData()` in _SearchbarMain.vue_

**If for some type of result you want to change the information / order of the information that the user sees in the corresponding rows of the result-suggestion list:**
  1. Locate this result type in record `TypeInfo`.
  2. In that entry, change / swap the references that are in field `howToFillresultSuggestionOutput`.

**If you want to change in depth the whole result-suggestion list (to change how every row displays the information):**
  1. Add a display-mode in the `SuggestionrowCells` enum in _searchbar.ts_
  2. Update the `SearchbarPurposeInfo` record there to tell the bar which Purpose must use your new mode.
  3. Implement this mode in a new root `<div>` at the end of the `<template>` of `SuggestionRow.vue`.

**You can create a new purpose if needed:**
  1. Add a purpose name into the `SearchbarPurpose` enum of _searchbar.ts_.
  2. Define the behavior of the bar when it has this purpose, by adding an entry in the `SearchbarPurposeInfo` record.
  3. Now you can give this purpose to the `:bar-purpose` props.

**If you want to add or remove a filter button:**
  - Either you simply need create or modify a purpose to see more/less filters (see above).
  - Or:
    1. Add/remove an entry in enum `Category`.
    2. Add/remove the corresponding category-title and button-label in enum `CategoryInfo`.
    3. You might need to add/remove an entry in `SubCategoryInfo`.
    4. Update (add/remove/change) all relevant entries in record `TypeInfo` to take properly into account your new categorization of the result types.
    5. Update the entries of `SearchbarPurposeInfo` to take into account the new/removed category.

**If you want to change the order of the results in the drop-down, it is a bit less straightforward:**
  - To change the order of the results inside each network set, modify the values of fields `priority` in the `TypeInfo` record of _searchbar.ts_.
  - Changing the order of the networks sets is a different task. You would need to change the `priority` fields in the `ChainInfo` record of _networks.ts_.

Note that if different types or networks have the same priority, two results of these types/networks will appear in the drop-down in the order of their `closeness` values
(a measure of similarity to the user input).

# Usage of _MiddleEllipsis.vue_

This component clips the text that you give to it. The text is clipped in the middle so the beginning and the end remain visible.

It has been designed to adapt the text to its width as quickly as possible (tests show that the optimization algorithm finds almost always the correct clipping within 3 iterations, often 2).

## Vocabulary

In the rest of the documentation, we will use those defintions:

- A component of _defined width_ is a component whose width does not collapse to `0px` or `min-width` when it has no content. So, for example, it has a `flex-grow` or `width` property set, or it is a cell in a grid and the width of its column is fixed in `px` or set to `auto` or `fr` with `grid-template-columns`.
- A component of _undefined width_ is a component whose width collapses to `0px` or `min-width` when it has no content.

## Syntax

### The simplest case

If the room allowed to the text is _defined_ (see the vocabulary above), you can write

```HTML
<MiddleEllipsis class="myclass" text="my long text" />
```
This works only when the room that the component has is independent of the room that other MiddleEllipsis components take on the same line.

### The interesting cases

**Coordination of interdependent MiddleEllipsis components**

In real applications, the simple case above is not always sufficient. You might need to display on the same line several MiddleEllipsis components whose widths are defined with `flex-grow` values,
which implies that the room that each component has depends on the room that other components take on the same line.
In this case, the simple syntax above will clip their texts wrongly. The reason is that they have no way to know that their widths depend on the content of all of them.

You must gather them in a parent MiddleEllipsis like so:
```HTML
<MiddleEllipsis class="papa">
  <MiddleEllipsis class="child1" text="a long text" />
  <MiddleEllipsis class="child2" text="another long text" />
  ...
</MiddleEllipsis>
```
Note that :
- A parent MiddleEllipsis can contain anything, so using a parent does not restrict the layout of your page. See it as a `div`. The parent displays components of other types as they are, unchanged. It recognizes and controls its children to make sure that they clip properly.
- Regarding the children, their display modes can be `inline-flex` or `inline-block` (only `inline` modes make sense in our context).

**Guaranteeing no gap between the clipped text and its neighbors**

When you want
- to have a room which adapts to its content (you are guaranteed that it is as large as its text, no matter how small the text is),
- but you do not want this room to exceed some limit (the text must clip at some point),

what width or flex-grow value should you use? There is no answer (unless the text is known in advance and fixed).
To achieve that, you can display your text through a MiddleEllipsis of undefined width as follows.

Take a parent of defined width. Inside, sit your MiddleEllipsis of undefined width and the other things to display on that line. For example:
```HTML
<MiddleEllipsis class="papa">
  <MiddleEllipsis class="child1" text="a long text" />
  <MiddleEllipsis class="undefwidth-child2" text="another long text" :initial-flex-grow="1" />
  ...
</MiddleEllipsis>
```
Note the use of props `initial-flex-grow`, which is mandatory for children of undefined width. It tells the component how much room it can give to its text with respect to the `flex-grow` properties of its neighbors. _initial_ means that the value gives room to clip the text during the computation. The real `flex-grow` of the component remains `0` so it collapses around its content, there is no empty space between it and its neighbors.

**Using more than one ellipsis to clip the text**

If you want your text to get clipped with a fixed number of ellipses, you can give a constant:
```HTML
<MiddleEllipsis class="myclass" text="my long text" :ellipses="3" />
```
MiddleEllipsis offers the possibility to adapt the number of ellipses to the length of the text. For example,
```HTML
<MiddleEllipsis class="myclass" text="my long text" :ellipses="[8,30]" />
```
tells the component to use one ellipsis if there is room for 8 characters or less, two ellipses between 9 and 30 characters, and three ellipses above. The array can configure as many cases as you want.

In practice, you will probably want a simple configuration like so:
```HTML
<MiddleEllipsis class="myclass" text="my long text" :ellipses="[8]" />
```
Here, the text will be clipped with one ellipsis if there is room for less than 9 characters, otherwise two ellipses.

## Restrictions

Never set a padding on the left or right side of a MiddleEllipsis component (margin is not a problem).

Never give a width to its left or right border.

If, somewhere in your CSS, some `@media (min-width: AAApx)` or `@media (max-width: AAApx)` queries have an (indirect) effect on the size of a MiddleEllipsis component,
you must inform the component that its width can jump unexpectedly. You do so by giving AAA to its `width-mediaquery-threshold` props.
For example
```HTML
  :width-mediaquery-threshold="600"
```
makes MiddleEllipsis aware that changes in the layout of the page happen when the width of the window/screen width passes through 600px.