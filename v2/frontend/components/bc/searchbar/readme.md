# Using the component in your code

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

# Changing the behavior of the search bar thanks to _searchbar.ts_

_searchbar.ts_ has been designed as a "configuration file" for the search bar. For example, if the protocol to communicate with the API changes, it might
not be necessary to modify the code of the search-bar (its behavior and look are hard-coded as little as possible).

If the API returns a new type of result:
  1. Add this type into the `ResultType` enum of _searchbar.ts_.
  2. Tell the bar how to read/display it by adding a new entry in the `TypeInfo` record.

If the API gets the ability to return a new field in some or all elements of its response array:
  1. Write the name of the new field in `SingleAPIresult`.
  2. Create a reference to it in `Indirect`.
  3. In `TypeInfo`, tell the bar when/where this field must be read (by giving its `Indirect` reference).
  4. Add a case for the reference in function `wasOutputDataGivenByTheAPI()`
  5. Add a case for the reference in function `realizeData()` in _SearchbarMain.vue_

If for some type of result you want to change the information / order of the information that the user sees in the corresponding rows of the result-suggestion list:
  1. Locate this result type in record `TypeInfo`.
  2. In that entry, change / swap the references that are in field `howToFillresultSuggestionOutput`.

If you want to change in depth the whole result-suggestion list (to change how every row displays the information):
  1. Add a display-mode in the `SuggestionrowCells` enum in _searchbar.ts_
  2. Update the `SearchbarPurposeInfo` record there to tell the bar which Purpose must use your new mode.
  3. Implement this mode in a new root `<div>` at the end of the `<template>` of `SuggestionRow.vue`.

You can create a new purpose if needed:
  1. Add a purpose name into the `SearchbarPurpose` enum of _searchbar.ts_.
  2. Define the behavior of the bar when it has this purpose, by adding an entry in the `SearchbarPurposeInfo` record.
  3. Now you can give this purpose to the `:bar-purpose` props.

If you want to add or remove a filter button:
  - Either you simply need create or modify a purpose to see more/less filters (see above).
  - Or:
    1. Add/remove an entry in enum `Category`.
    2. Add/remove the corresponding category-title and button-label in enum `CategoryInfo`.
    3. You might need to add/remove an entry in `SubCategoryInfo`.
    4. Update (add/remove/change) all relevant entries in record `TypeInfo` to take properly into account your new categorization of the result types.
    5. Update the entries of `SearchbarPurposeInfo` to take into account the new/removed category.

If you want to change the order of the results in the drop-down, it is a bit less straightforward:
  - To change the order of the results inside each network set, modify the values of fields `priority` in the `TypeInfo` record of _searchbar.ts_.
  - Changing the order of the networks sets is a different task. You would need to change the `priority` fields in the `ChainInfo` record of _networks.ts_.
  Note that if different types or networks have the same priority, two results of these types/networks will appear in the drop-down in the order of their `closeness` values
  (a measure of similarity to the user input).

