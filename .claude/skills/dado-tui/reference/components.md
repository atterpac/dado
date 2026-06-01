# Component Catalog

Components live in `components/*.go` (package `components`) except **CommandBar** (`input/commandbar.go`, package `input`). Most builders are fluent (return the receiver). Tiers mirror `cmd/tutorial/demos/{basic,intermediate,advanced}/`.

Several "components" share a file: `primitives.go` (Badge, Chip, Divider, Skeleton), `progress.go` (ProgressBar, Spinner, Gauge, Sparkline), `select.go` (Select, MultiSelect), `checkbox.go` (Checkbox, RadioGroup), `textfield.go` (TextField, TextArea), `composites.go` (StatusBar, SearchBar).

---

## Basic

| Component | Constructor | Key methods | Use when |
|---|---|---|---|
| **Badge** | `NewBadge(text)` | `SetText`, `SetVariant`, `SetPill(bool)`, `SetIcon` | inline status pill, count, tag |
| **Checkbox** | `NewCheckbox(name)` | `SetLabel`, `SetChecked`, `SetOnChange(ChangeHandler[bool])` | single on/off setting |
| **Chip** | `NewChip(text)` | `SetIcon`, `SetRemovable`, `SetSelected`, `SetOnRemove`, `SetOnClick` | filter tags, dismissable tokens |
| **Divider** | `NewDivider()` / `NewVerticalDivider()` | `SetLabel`, `SetStyle(rune)`, `SetOrientation` | section a layout |
| **EmptyState** | `NewEmptyState()` | `SetIcon`, `SetTitle`, `SetMessage` | a list/table/view has no data |
| **MultiSelect** | `NewMultiSelect(name)` | `SetLabel`, `SetOptions`/`SetOptionsWithValues`, `SetSelected([]int)`, `SetOnChange` | multi-value form field |
| **Panel** | `NewPanel()` | `SetContent(primitive)`, `SetTitle`, `SetTitleColor`, `SetTitleAlign`, `SetFocused` | frame/label any child |
| **ProgressBar** | `NewProgressBar()` | `SetProgress(float64)`, `SetLabel`, `SetShowPercentage`, `SetChars` | determinate progress |
| **RadioGroup** | `NewRadioGroup(name)` | `SetLabel`, `SetOptions`, `SetSelected(int)`, `SetOnChange` | one-of-N form selection |
| **Select** | `NewSelect(name)` | `SetLabel`, `SetPlaceholder`, `SetOptions`/`SetOptionsWithValues`, `SetDefault`, `SetOnChange` | compact one-of-N input |
| **Skeleton** | `NewSkeleton()` | `SetVariant`, `SetLines(int)`, `SetAnimated` | loading placeholder shimmer |
| **Spinner** | `NewSpinner()` | `SetStyle`, `SetLabel`, `SetInterval(d)` | indeterminate activity |
| **Tabs** | `NewTabs()` | `AddTab(name,content)`, `AddTabWithIcon`, `SetActive`/`SetActiveByName`, `SetBadge`, `SetClosable`, `SetOnChange`/`SetOnClose` | switch views in one pane |
| **TextArea** | `NewTextArea(name)` | `SetLabel`, `SetPlaceholder`, `SetValue`, `SetMaxLines`, `SetOnChange` | multi-line text entry |
| **TextField** | `NewTextField(name)` | `SetLabel`, `SetPlaceholder`, `SetMasked`/`SetMaskChar`, `SetValue`, `SetValidators`, `SetOnChange`/`SetOnSubmit` | short input, password, validated |
| **Button** | `NewButton(label)` | `SetLabel`, `SetVariant`, `SetDisabled`, `OnClick(func())` | trigger an action |
| **Label** | `NewLabel(text)` | `SetText`, `SetAlign`, `SetColor`, `SetBold`, `SetWordWrap`, `SetScrollable`, `SetDynamicColors` | heading, caption, read-only text |
| **ToastManager** | `NewToastManager(app)` | `Info`/`Success`/`Warning`/`Error(msg)`, `ShowWithUndo`, `ShowWithAction`, `ShowPersistent`; config `SetPosition`/`SetMaxVisible`/`SetDefaultDuration` | transient post-action feedback |

---

## Intermediate

| Component | Constructor | Key methods | Use when |
|---|---|---|---|
| **BarChart** | `NewBarChart()` | `SetItems(...BarItem)`/`SetValues`, `SetOrientation`, `SetRange`/`SetAutoScale`, `SetShowValues`, `SetOnSelect` | compare discrete categories |
| **CodeView** | `NewCodeView()` | `SetCode`/`SetLines`, `SetLanguage`, `SetShowLineNumbers`, `SetTabWidth`, `SetWrapLines`, `SetHighlightLine`, `SetOnLineClick` | syntax-highlighted source |
| **Gauge** | `NewGauge()` | `SetValue`, `SetLabel`, `SetUnit`, `SetMaxValue` | one metric vs a max (CPU %) |
| **HeatMap** | `NewHeatMap()` | `SetData([][]HeatMapCell)`/`SetValues`, `SetRowLabels`/`SetColLabels`, `SetColorScale`, `SetShowValues`, `SetOnSelect`/`SetOnHover` | matrices, activity calendars |
| **LineGraph** | `NewLineGraph()` | `SetSeries(...DataSeries)`/`AddValue`, `SetRange`/`SetAutoScale`, `SetShowLegend`, `SetXAxis`/`SetYAxis`, `SetShowGrid`, `SetOnHover` | trends over time |
| **LogViewer** | `NewLogViewer()` | `AddEntry`, `SetMaxEntries`, `SetShowTimestamp/Level/Source`, `SetFollow`, `SetMinLevel`, `SetSearch`/`SetSearchRegex`, `SetFilter`, `SetOnSelect` | live tailing logs w/ filter+search |
| **Modal** | `NewModal(ModalConfig)` | `SetContent`, `SetHints`, `SetOnSubmit`/`SetOnCancel`/`SetOnClose`, `SetDismissOnEsc`, `SetFocusOnShow` | confirmations, focused dialogs (presets in `modal_presets.go`) |
| **Sparkline** | `NewSparkline()` | `SetValues`/`AddValue`, `SetMaxValue`, `SetLabel` | compact micro-trend in cards |
| **Split** | `NewSplit()` | `SetDirection`, `SetRatio(float64)`, `SetFirst`/`SetSecond` (or `SetLeft/Right/Top/Bottom`), `SetResizable`, `SetShowDivider`, `SetOnResize` | side-by-side / stacked resizable |
| **Table** | `NewTable()` | `SetHeaders`, `AddRow`/`AddColoredRow`/`AddStyledRow`, `SetMultiSelect`, `SetOnSelect`/`SetOnSelectionChange`, `SetEmptyTitle/Message` | structured tabular data w/ selection |
| **Timeline** | `NewTimeline()` | `SetLanes([]TimelineLane)`/`AddLane`, `SetLabelWidth`, `SetShowLegend`, `SetTimeRange`, `SetZoom`, `SetOnSelect` | schedules, Gantt-style bars |
| **Tree** | `NewTree()` | `SetRoot(*TreeNode)`, `SetShowLines`, `SetShowIcons`, `SetMultiSelect`, `SetLazyLoader`, `SetOnSelect`/`SetOnExpand`/`SetOnCollapse` | file trees, nested hierarchies |
| **VirtualList** | `NewVirtualList()` | `SetItems`/`SetTotalCount`+`SetFetchFunc`, `SetRenderFunc`, `SetOverscan`, `SetPageSize`, `SetShowScrollbar`, `SetOnSelect`/`SetOnScrollEnd` | thousands of rows / paginated data |

---

## Advanced

| Component | Constructor | Key methods | Use when |
|---|---|---|---|
| **Autocomplete** | `NewAutocompleteInput()` | `SetPrompt`, `SetPlaceholder`, `SetMaxSuggestions`, `SetSuggestionProvider`, `SetHistoryProvider`, `SetOnSubmit`/`SetOnSelect` | command/search input w/ completion |
| **CommandBar** | `input.NewCommandBar()` | `SetInput`, `SetOnSubmit(func(cmdType, input))`, `SetOnCancel`, `SetOnChange` | `:`-style command/mode bar |
| **ContextMenu** | `NewContextMenu()` | `AddItem(id,label,handler)`, `AddItemWithShortcut/Icon`, `AddSubmenu`, `AddDivider`, `AddSection`, `SetDisabled`/`SetChecked`/`SetDanger`, `SetOnSelect`/`SetOnClose` | contextual action lists, submenus |
| **DiffViewer** | `NewDiffViewer()` | `SetDiff(old,new)`/`SetUnifiedDiff`/`SetDiffResult`, `SetSideBySide`, `SetShowLineNumbers`, `SetWordDiff`, `SetContextLines`, `SetOnLineSelect`/`SetOnHunkAction` | show code/text changes |
| **Drawer** | `NewDrawer(DrawerConfig)` | `SetContent`, `SetHints`, `SetBehavior`, `SetDismissOnEsc`, `SetOnClose`/`SetOnDismiss` | side panels (details, settings, nav) |
| **Finder** | `NewFinder()` | `SetItems`/`SetCategories`, `SetPlaceholder`/`SetPrompt`, `SetMinScore`, `SetPreview(fn)`/`SetPreviewRatio`, `SetVimMode`, `SetRecentItems`, `SetOnSelect`/`SetOnChange` | fzf-style fuzzy search/jump |
| **Form** | `NewForm()` | `AddTextField`/`AddSelect`/`AddCheckbox`/`AddRadioGroup`/`AddField`, `SetValues(map)`, `SetOnSubmit(func(map[string]any))`, `SetOnCancel` | quick data-entry form |
| **FormBuilder** | `NewFormBuilder()` | `Text`/`TextArea`/`Select`/`SelectWithValues`/`MultiSelect`/`Checkbox`/`Radio` (each → field builder w/ `OnChange`/`Validate`/`Done`), `AddField`, `OnSubmit`, `OnCancel`, `Build()` | fluent form w/ per-field config + validation |
| **GitGraph** | data: `NewGitGraphData()`+`AddCommit`; widget: `NewGitGraph()` | `SetGraph(data)`, `SetShowRefs/Hash/Author/Date`, `SetDateFormat`, `SetLaneColors`, `SetOnSelect`/`SetOnChange` | git history w/ branch lanes |
| **MasterDetail** | `NewMasterDetailView()` / `NewMasterDetailViewConfig(cfg)` | `SetMasterContent`/`SetDetailContent`, `SetMasterTitle`/`SetDetailTitle`, `SetRatio`, `SetResizable`, `SetDetailVisible`, `SetOnSelectionChange`/`SetOnDetailToggle` | browse-then-inspect (email/files) |
| **MetricCard** | `NewMetricCard()` | `SetLabel`, `SetValue`/`SetNumericValue`, `SetUnit`, `SetTrend`, `SetSparkline([]float64)`/`AddSparkValue`, `SetThresholds`, `SetCompact` | dashboard KPI w/ trend + sparkline |
| **ProgressModal** | `NewProgressModal()` | `SetTitle`, `SetCancelable`, `SetIndeterminate`, `SetProgress(float64)`, `SetMessage`/`SetSubMessage`, `SetOnCancel`/`SetOnComplete`/`SetOnClose` | block UI during a long op |
| **SearchBar** | `NewSearchBar()` | `SetPlaceholder`, `SetIcon`, `SetResults([]SearchResult)`, `SetShowSpinner`, `SetMaxResults`, `SetOnSearch`/`SetOnSelect`/`SetOnChange`/`SetOnCancel` | search-as-you-type w/ results list |
| **Splash** | `NewSplash()` | `SetLogo`, `SetStatus`, `SetGradient`/`SetColors`, `SetAutoDismiss(d)`, `SetDismissKeys`, `SetDevMode`, `SetOnClose` | startup/loading screen |
| **StatusBar** (component) | `NewStatusBar()` | `SetLeft`/`SetCenter`/`SetRight(...StatusSection)`, `AddLeft`/`AddRight`, `SetSeparator`, `SetShowBorder` | persistent status / hint line |
| **DataGrid** | `NewDataGrid()` | `SetSource(DataGridSource)`, `SetShowRowNumbers`/`SetShowHeader`, `SetOverscan`, `SetOnCellEdit`/`SetOnModalEdit`, `SetOnCellSelect`/`SetOnModeChange`/`SetOnChangesetUpdate`/`SetOnSubmit` | large editable grid, vim modes, changesets |
| **ERDGraph** | `NewERDGraph()` | `SetData(tables, relations)`, `SetNodeWidth`, `SetSpacing(h,v)`, `SetFocusedTable`, `SetOnSelect(func(*ERDTable))` | DB schema / relationships |
| **NodeGraph** | data: `NewNodeGraphData()`+`AddNode`/`AddEdge`; widget: `NewNodeGraph()` | `SetData`, `SetNodeWidth`, `SetShowEdgeLabels`, `SetFocus(nodeID)`, `SetOnSelect` | dependency/flow/network diagram |
| **GraphTree** | data: `NewGraphTreeData()`+`AddNode`/`AddEdge`; widget: `NewGraphTree()` | `SetData`, `SetShowEdgeLabels`, `SetOnLoadChildren(fn)`, `SetOnSelect`/`SetOnChange` | DAG/hierarchy w/ lazy expand |
| **BottomSheet** | `NewBottomSheet(BottomSheetConfig)` | `SetContent`, `SetHeight`, `SetHints`, `SetBehavior`, `SetDismissOnEsc`, `SetOnClose`/`SetOnDismiss` | mobile-style bottom action/detail sheet |
| **HintGrid** | `NewHintGrid()` | `SetHints([]KeyHint)` | help overlay / shortcut cheat-sheet |

---

## Patterns across the catalog

- **Overlay family** (`Modal`, `Drawer`, `BottomSheet`) share a `ModalBehavior` + dismiss lifecycle (implement `nav.ModalComponent`) and take a `*Config` struct.
- **Graph family** (`GitGraph`, `NodeGraph`, `GraphTree`, `ERDGraph`) use a data-object + view pattern: build a `*Data` object with `AddNode`/`AddEdge`/`AddCommit`, then `SetData`/`SetGraph` on the widget.
- **Multi-file components**: `DataGrid` (`datagrid_*.go`) and `ERDGraph` (`erdgraph_*.go`) split render/input/layout; the type + constructor + main setters live in the base file.
- **Form fields** all implement the value/validation interfaces (`reference/architecture.md`), so `FormBuilder` and `binding.FormBinding` treat them uniformly.
- **Toast** is a manager (`ToastManager`), not a single widget — call `Info/Success/Warning/Error/ShowWithUndo`.
