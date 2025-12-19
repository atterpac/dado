package theme

// Icon constants using Nerd Font glyphs.
// See https://www.nerdfonts.com/cheat-sheet for full list.

const (
	// Status icons
	IconCheck    = "\uf00c" // nf-fa-check
	IconError    = "\uf00d" // nf-fa-times
	IconWarning  = "\uf071" // nf-fa-exclamation_triangle
	IconInfo     = "\uf05a" // nf-fa-info_circle
	IconPending  = "\uf10c" // nf-fa-circle_o
	IconRunning  = "\uf144" // nf-fa-play_circle
	IconCanceled = "\uf05e" // nf-fa-ban
	IconSkipped  = "\uf04e" // nf-fa-forward

	// Navigation arrows
	IconArrowRight = "\uf054" // nf-fa-chevron_right
	IconArrowLeft  = "\uf053" // nf-fa-chevron_left
	IconArrowUp    = "\uf077" // nf-fa-chevron_up
	IconArrowDown  = "\uf078" // nf-fa-chevron_down
	IconChevronR   = "\uf105" // nf-fa-angle_right
	IconChevronL   = "\uf104" // nf-fa-angle_left
	IconChevronU   = "\uf106" // nf-fa-angle_up
	IconChevronD   = "\uf107" // nf-fa-angle_down

	// Tree/hierarchy
	IconTreeBranch = "\u251c" // ├
	IconTreeLast   = "\u2514" // └
	IconTreeVert   = "\u2502" // │
	IconTreeHoriz  = "\u2500" // ─
	IconTreeEmpty  = " "

	// Box drawing (rounded corners)
	IconCornerTL = "\u256d" // ╭
	IconCornerTR = "\u256e" // ╮
	IconCornerBL = "\u2570" // ╰
	IconCornerBR = "\u256f" // ╯
	IconHorizBar = "\u2500" // ─
	IconVertBar  = "\u2502" // │

	// Box drawing (square corners)
	IconCornerTLSquare = "\u250c" // ┌
	IconCornerTRSquare = "\u2510" // ┐
	IconCornerBLSquare = "\u2514" // └
	IconCornerBRSquare = "\u2518" // ┘

	// UI elements
	IconFolder     = "\uf07b" // nf-fa-folder
	IconFolderOpen = "\uf07c" // nf-fa-folder_open
	IconFile       = "\uf15b" // nf-fa-file
	IconFileCode   = "\uf1c9" // nf-fa-file_code_o
	IconSearch     = "\uf002" // nf-fa-search
	IconFilter     = "\uf0b0" // nf-fa-filter
	IconRefresh    = "\uf021" // nf-fa-refresh
	IconCopy       = "\uf0c5" // nf-fa-copy
	IconEdit       = "\uf044" // nf-fa-edit
	IconDelete     = "\uf1f8" // nf-fa-trash
	IconAdd        = "\uf067" // nf-fa-plus
	IconClose      = "\uf00d" // nf-fa-times
	IconMenu       = "\uf0c9" // nf-fa-bars
	IconSettings   = "\uf013" // nf-fa-cog
	IconHome       = "\uf015" // nf-fa-home
	IconList       = "\uf03a" // nf-fa-list
	IconGrid       = "\uf00a" // nf-fa-th

	// Connection
	IconConnected    = "\uf1e6" // nf-fa-plug
	IconDisconnected = "\uf127" // nf-fa-chain_broken
	IconCloud        = "\uf0c2" // nf-fa-cloud
	IconServer       = "\uf233" // nf-fa-server
	IconDatabase     = "\uf1c0" // nf-fa-database

	// Time
	IconClock    = "\uf017" // nf-fa-clock_o
	IconCalendar = "\uf073" // nf-fa-calendar
	IconHistory  = "\uf1da" // nf-fa-history
	IconTimer    = "\uf254" // nf-fa-hourglass_half

	// Actions
	IconPlay   = "\uf04b" // nf-fa-play
	IconPause  = "\uf04c" // nf-fa-pause
	IconStop   = "\uf04d" // nf-fa-stop
	IconReplay = "\uf01e" // nf-fa-rotate_right
	IconSave   = "\uf0c7" // nf-fa-save
	IconLoad   = "\uf093" // nf-fa-upload
	IconExport = "\uf019" // nf-fa-download
	IconImport = "\uf093" // nf-fa-upload

	// Misc
	IconStar     = "\uf005" // nf-fa-star
	IconHeart    = "\uf004" // nf-fa-heart
	IconBolt     = "\uf0e7" // nf-fa-bolt
	IconFire     = "\uf06d" // nf-fa-fire
	IconBug      = "\uf188" // nf-fa-bug
	IconKey      = "\uf084" // nf-fa-key
	IconLock     = "\uf023" // nf-fa-lock
	IconUnlock   = "\uf09c" // nf-fa-unlock
	IconUser     = "\uf007" // nf-fa-user
	IconUsers    = "\uf0c0" // nf-fa-users
	IconTag      = "\uf02b" // nf-fa-tag
	IconBookmark = "\uf02e" // nf-fa-bookmark

	// Separators / decorative
	IconDot       = "\uf111" // nf-fa-circle
	IconBullet    = "\uf192" // nf-fa-dot_circle_o
	IconDiamond   = "\u25c6" // ◆
	IconTriangleR = "\u25b6" // ▶
	IconTriangleL = "\u25c0" // ◀
	IconTriangleU = "\u25b2" // ▲
	IconTriangleD = "\u25bc" // ▼
	IconBlock     = "\u2588" // █
	IconBlockHalf = "\u258c" // ▌

	// Domain-specific icons (Temporal workflows, etc.)
	IconWorkflow   = "\uf0e7" // nf-fa-bolt - represents workflow execution flow
	IconActivity   = "\uf013" // nf-fa-cog - represents activity execution
	IconTaskQueue  = "\uf0ae" // nf-fa-tasks - represents task queue
	IconEvent      = "\uf1da" // nf-fa-history - represents an event
	IconSignal     = "\uf012" // nf-fa-signal - represents workflow signal
	IconSchedule   = "\uf073" // nf-fa-calendar - represents schedule
	IconTimedOut   = "\uf017" // nf-fa-clock_o - represents timeout
	IconCompleted  = "\uf00c" // nf-fa-check - alias for IconCheck
	IconFailed     = "\uf00d" // nf-fa-times - alias for IconError
	IconTerminated = "\uf28d" // nf-fa-stop_circle
	IconNamespace  = "\uf0e8" // nf-fa-sitemap

	// Tree view icons
	IconTreeExpanded  = "\uf0d7" // nf-fa-caret_down
	IconTreeCollapsed = "\uf0da" // nf-fa-caret_right
	IconTreeLeaf      = "\uf111" // nf-fa-circle (small dot)

	// Progress bar
	IconBarFull    = "\u2588" // █
	IconBarHalf    = "\u2584" // ▄
	IconBarEmpty   = "\u2591" // ░
	IconBarRunning = "\u2593" // ▓
)
