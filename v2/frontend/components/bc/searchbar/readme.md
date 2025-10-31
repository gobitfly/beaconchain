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
  :bar-shape="SearchbarShape.<what you want to see>"
  :color-theme="SearchbarColors.<theme determining the colors of the bar and in the dropdown>"
  :bar-purpose="SearchbarPurpose.<what you want the bar to do>"
  :screen-width-causing-sudden-change="<number>"
  :pick-by-default="<function that picks a result on behalf of the user when they press Enter or the search-button>"
  @go="<your function doing something when a result is selected>"
/>
```

There are more props that you can give to configure the search bar:

```TS
:only-networks="[<list of chain IDs that the bar is authorized to search over>]" // Without this prop, the bar searches over all networks.
:keep-dropdown-open="true" // When the user selects a result, the drop-down does not close.
:row-lacks-premium-subscription="<call-back function returning `true` if the result-suggestion that it is passed must be deactivated>" // For the rows that the function returns `true` on, the user is invited to subscribe to a premium plan
```

The list of possible values for `:bar-shape`, `:color-theme` and `:bar-purpose` are respectively in enums `SearchbarShape`, `SearchbarColors` and `SearchbarPurpose` in file _searchbar.ts_.

If the width of the search bar can change suddenly while the user is redimensionning the window (for example due to a media query in the CSS or some JS code), the width threshold (in pixels) at which you trigger the change must be passed in prop `:screen-width-causing-sudden-change`. This prevents visual bugs in the list of results (to understand why, see section "Sudden changes of width" in the documentation of _MiddleEllipsis.vue_).

You can write your own function for `:pick-by-default`, or give the example function written in _searchbar.ts_ if it suits your needs:

```TS
:pick-by-default="pickHighestPriorityAmongBestMatchings"
```

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

1. Write the name of the new field in `SingleAPIresult` in _searchbar.ts_.
2. Create a reference to it in `Indirect`.
3. In `TypeInfo`, tell the bar when/where this field must be read (by giving its `Indirect` reference).
4. Add a case for the reference in function `wasOutputDataGivenByTheAPI()`.
5. Add a case for the reference in function `realizeData()`.

**If for some type of result you want to change the information / order of the information that the user sees in the corresponding rows of the result-suggestion list:**

1. Locate this result type in record `TypeInfo`.
2. In that entry, change / swap the references that are in field `howToFillresultSuggestionOutput`.

**If you want to change in depth the whole result-suggestion list (to change how every row displays the information):**

1. Add a display-mode in the `SuggestionrowCells` enum in _searchbar.ts_.
2. Update the `SearchbarPurposeInfo` record there to tell the bar which Purpose must use your new mode.
3. Implement this mode in a new root `<div>` at the end of the `<template>` of `SuggestionRow.vue`.

**You can create a new purpose if needed:**

1. Add a purpose name into the `SearchbarPurpose` enum of _searchbar.ts_.
2. Define the behavior of the bar when it has this purpose, by adding an entry in the `SearchbarPurposeInfo` record.
3. Now you can give this purpose to the `:bar-purpose` prop.

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

In the rest of the documentation, we will use these definitions:

- A component of _defined width_ is a component whose width does not collapse to `0px` or `min-width` when it has no content. So, for example, it has a `flex-grow` or `width` property set, or it is a cell in a grid and the width of its column is fixed in `px` or set to `auto` or `fr` with `grid-template-columns`, ...
- A component of _undefined width_ is a component whose width collapses to `0px` or `min-width` when it has no content.

## Syntax

### The simplest case

If the width of the component is _defined_ (see the vocabulary above) and its width is independent of the content of other MiddleEllipsis components, you can write

```HTML
<MiddleEllipsis class="myclass" text="my long text" />
```

When the width of the component is undefined or depends on the content of other MiddleEllipses around, the syntax is different, as we will explain now:

### The interesting cases

**Coordination of interdependent MiddleEllipsis components**

In real applications, you might need to display on the same line several MiddleEllipsis components whose widths are defined with `flex-grow` values,
which implies that the room that each component has depends on the content of the others (because the texts push or pull their containers depending on their relative lengths).

In that case, the simple syntax above causes 2 problems:

1. The page can become slow and freeze. The reason is that each reclipping in a component changes its width, so the width of its neighbors, thus triggering them, which initiates an infinite loop of updates.
2. It will clip the texts wrongly. The reason is that each component does not know the final widths of its neighbors while they are clipping independently.

So, in that case, you must give the components a way to deal properly with each other. You do it by gathering them in a parent MiddleEllipsis like so:

```HTML
<MiddleEllipsis class="papa">
  <MiddleEllipsis class="child1" text="a long text" />
  <MiddleEllipsis class="child2" text="another long text" />
  ...
</MiddleEllipsis>
```

Note that :

- A parent MiddleEllipsis can contain anything, so using a parent does not restrict the layout of your page. See it as a `div`. It recognizes and manages its children to make sure that they clip properly, and displays components of other types as they are, unchanged.
- Regarding the children, their display modes can mix `inline-flex` and `inline-block` (only `inline` modes make sense in our context).

**Guaranteeing no gap between the clipped text and its neighbors: using the concept of _undefined width_**

When you want

- a MiddleEllipsis component to be as narrow as its text (no empty space), no matter how small the text is,
- and prevent it from exceeding some limit when the text is large (the text must clip at some point),

then you can display your text through a MiddleEllipsis of undefined width as follows.

Take a parent of defined width. Inside, sit your MiddleEllipsis of undefined width and the other things to display on the same line. For example:

```HTML
<MiddleEllipsis class="papa">
  blabla bla, look there is no gap:
  <MiddleEllipsis class="undefwidth-child" text="a long text" :initial-flex-grow="1" />
  immediately followed by the rest without empty space
</MiddleEllipsis>
```

Note

- the use of prop `initial-flex-grow`, which is mandatory for children of undefined width. It tells the component how much room it can give to its text with respect to the `flex-grow` properties of its neighbors.
  _initial_ means that the value defines the room during the computation only, the real `flex-grow` of the component remains `0` so the component collapses around its content, there is no empty space between it and its neighbors.
- You can of course put other things in the parent, for example other MiddleEllipses of defined or undefined widths.

**Using more than one ellipsis to clip the text**

If you want your text to get clipped with a certain number of ellipses, you can give a constant:

```HTML
<MiddleEllipsis class="myclass" text="my long text" :ellipses="3" />
```

MiddleEllipsis offers to adapt the number of ellipses to the length of the text. For example,

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

MiddleEllipsis cannot detect and react when the font of the text changes. It does not forbid you to change the font, but it restricts the way you can do it:
If you change the font while the component is being resized, it is fine because the component checks the font before each reclipping.
If you change the font independently of the width of the component, then perform the change by swapping a class of the component because it detects changes in its list of classes and reclips.

### Sudden changes of width

When the user resizes her/his window, thus changing the width of a MiddleEllipsis component, the component adjusts its clipping automatically.

But your website might switch between layouts (for example between desktop and mobile modes) when the width of the window reaches a certain threshold. This reorganizes the components and changes their sizes suddenly. This is typically performed in CSS by some `@media (min-width: ...)` or `@media (max-width: ...)` query, or in JS.

You must inform MiddleEllipsis that its width jumps unexpectedly. Othewise, it might not detect it (the text will overflow or waste blank space).

There are two ways to pass this information. Either

- You give the threshold at which you perform the change to its `width-mediaquery-threshold` prop.
  For example, `:width-mediaquery-threshold="600"` makes MiddleEllipsis aware that changes in the layout of the page happen when the width of the window/screen passes through 600px.
- Or you remove/add/swap a class in the class-list of the component at the very moment the component has its width modified. The fact that the list of classes changes makes the component reclip.

Children do not need this information, only parents and stand-alone components.
