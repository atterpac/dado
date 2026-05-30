package async

import (
	"fmt"
	"sync"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
	"github.com/rivo/tview"
)

// LoadingIndicator defines the interface for loading indicators.
// Implement this interface to create custom loading indicators.
type LoadingIndicator interface {
	// Show displays the loading indicator
	Show()

	// Hide removes the loading indicator
	Hide()

	// Success is called when the operation succeeds (before Hide)
	Success()

	// Error is called when the operation fails (before Hide)
	Error(err error)
}

// IndicatorConfig provides common configuration for built-in indicators.
type IndicatorConfig struct {
	// Message shown while loading
	Message string

	// SuccessMessage shown briefly on success (optional)
	SuccessMessage string

	// ShowSuccess controls whether to show success feedback
	ShowSuccess bool

	// ShowError controls whether to show error feedback
	ShowError bool
}

// DefaultConfig returns a default indicator configuration.
func DefaultConfig(message string) IndicatorConfig {
	return IndicatorConfig{
		Message:        message,
		SuccessMessage: "Done",
		ShowSuccess:    false, // Don't show success by default (cleaner UX)
		ShowError:      true,  // Always show errors
	}
}

// --- Toast Indicator ---

// toastIndicator shows loading state via ToastManager.
type toastIndicator struct {
	config  IndicatorConfig
	manager *components.ToastManager
	toast   *components.Toast
	mu      sync.Mutex
}

// Toast creates a loading indicator that shows a toast notification.
// The toast appears during loading and auto-dismisses on completion.
//
// Example:
//
//	async.NewLoader[Data]().
//	    WithIndicator(async.Toast("Loading data...")).
//	    Run(fetchData)
func Toast(message string) LoadingIndicator {
	return ToastWithConfig(DefaultConfig(message))
}

// ToastWithConfig creates a toast indicator with custom configuration.
func ToastWithConfig(config IndicatorConfig) LoadingIndicator {
	return &toastIndicator{
		config: config,
	}
}

// ToastWithManager creates a toast indicator using a specific ToastManager.
// Use this if you have a custom ToastManager instance.
func ToastWithManager(manager *components.ToastManager, message string) LoadingIndicator {
	return &toastIndicator{
		config:  DefaultConfig(message),
		manager: manager,
	}
}

func (t *toastIndicator) Show() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get or create toast manager
	if t.manager == nil {
		app := theme.GetApp()
		if app == nil {
			return
		}
		t.manager = components.NewToastManager(app)
	}

	// Show persistent loading toast
	t.toast = t.manager.ShowPersistent(t.config.Message, components.ToastInfo)
}

func (t *toastIndicator) Hide() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.manager != nil && t.toast != nil {
		t.manager.Dismiss(t.toast.ID)
		t.toast = nil
	}
}

func (t *toastIndicator) Success() {
	if !t.config.ShowSuccess {
		return
	}

	t.mu.Lock()
	manager := t.manager
	t.mu.Unlock()

	if manager != nil {
		manager.Success(t.config.SuccessMessage)
	}
}

func (t *toastIndicator) Error(err error) {
	if !t.config.ShowError {
		return
	}

	t.mu.Lock()
	manager := t.manager
	t.mu.Unlock()

	if manager != nil {
		manager.Error(err.Error())
	}
}

// --- Status Bar Indicator ---

// StatusFunc is a function that updates a status message.
type StatusFunc func(message string)

// statusIndicator shows loading state via a status bar or message area.
type statusIndicator struct {
	config    IndicatorConfig
	setStatus StatusFunc
	mu        sync.Mutex
}

// StatusBar creates a loading indicator that updates a status bar/message.
// Provide a function that sets the status message in your UI.
//
// Example:
//
//	async.NewLoader[Data]().
//	    WithIndicator(async.StatusBar("Loading...", app.SetStatusMessage)).
//	    Run(fetchData)
func StatusBar(message string, setStatus StatusFunc) LoadingIndicator {
	return &statusIndicator{
		config:    DefaultConfig(message),
		setStatus: setStatus,
	}
}

// StatusBarWithConfig creates a status bar indicator with custom configuration.
func StatusBarWithConfig(config IndicatorConfig, setStatus StatusFunc) LoadingIndicator {
	return &statusIndicator{
		config:    config,
		setStatus: setStatus,
	}
}

func (s *statusIndicator) Show() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.setStatus != nil {
		s.setStatus(s.config.Message)
	}
}

func (s *statusIndicator) Hide() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.setStatus != nil {
		s.setStatus("")
	}
}

func (s *statusIndicator) Success() {
	if !s.config.ShowSuccess {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.setStatus != nil {
		s.setStatus(s.config.SuccessMessage)
	}
}

func (s *statusIndicator) Error(err error) {
	if !s.config.ShowError {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.setStatus != nil {
		s.setStatus(fmt.Sprintf("Error: %v", err))
	}
}

// --- Progress Modal Indicator ---

// ModalShowFunc is a function that displays a modal.
type ModalShowFunc func(modal *components.ProgressModal)

// ModalHideFunc is a function that hides a modal.
type ModalHideFunc func()

// progressIndicator shows loading state via a ProgressModal.
type progressIndicator struct {
	config IndicatorConfig
	modal  *components.ProgressModal
	show   ModalShowFunc
	hide   ModalHideFunc
	mu     sync.Mutex
}

// ProgressModal creates a loading indicator that shows a progress modal.
// Provide functions to show/hide the modal in your UI.
//
// Example:
//
//	async.NewLoader[Data]().
//	    WithIndicator(async.ProgressModal(
//	        "Loading",
//	        "Fetching data...",
//	        func(m *components.ProgressModal) { pages.AddPage("loading", m, true, true) },
//	        func() { pages.RemovePage("loading") },
//	    )).
//	    Run(fetchData)
func ProgressModal(title, message string, show ModalShowFunc, hide ModalHideFunc) LoadingIndicator {
	return &progressIndicator{
		config: IndicatorConfig{
			Message:   message,
			ShowError: true,
		},
		show: show,
		hide: hide,
	}
}

func (p *progressIndicator) Show() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.modal = components.NewProgressModal().
		SetTitle(p.config.Message).
		SetMessage(p.config.Message).
		SetIndeterminate(true)

	// Start spinner animation
	if app := theme.GetApp(); app != nil {
		p.modal.StartSpinner(app)
	}

	if p.show != nil {
		p.show(p.modal)
	}
}

func (p *progressIndicator) Hide() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.hide != nil {
		p.hide()
	}
	p.modal = nil
}

func (p *progressIndicator) Success() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.modal != nil {
		p.modal.Complete("Done")
	}
}

func (p *progressIndicator) Error(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.modal != nil {
		p.modal.Fail(err)
	}
}

// --- Callback Indicator ---

// callbackIndicator wraps custom show/hide functions.
type callbackIndicator struct {
	onShow    func()
	onHide    func()
	onSuccess func()
	onError   func(error)
}

// Callback creates a custom indicator from callback functions.
// Use this for complete control over loading state display.
//
// Example:
//
//	async.NewLoader[Data]().
//	    WithIndicator(async.Callback(
//	        func() { spinner.Show() },
//	        func() { spinner.Hide() },
//	    )).
//	    Run(fetchData)
func Callback(onShow, onHide func()) LoadingIndicator {
	return &callbackIndicator{
		onShow: onShow,
		onHide: onHide,
	}
}

// CallbackFull creates a custom indicator with all callback options.
func CallbackFull(onShow, onHide func(), onSuccess func(), onError func(error)) LoadingIndicator {
	return &callbackIndicator{
		onShow:    onShow,
		onHide:    onHide,
		onSuccess: onSuccess,
		onError:   onError,
	}
}

func (c *callbackIndicator) Show() {
	if c.onShow != nil {
		c.onShow()
	}
}

func (c *callbackIndicator) Hide() {
	if c.onHide != nil {
		c.onHide()
	}
}

func (c *callbackIndicator) Success() {
	if c.onSuccess != nil {
		c.onSuccess()
	}
}

func (c *callbackIndicator) Error(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}

// --- Primitive Indicator ---

// primitiveIndicator shows/hides a tview primitive.
type primitiveIndicator struct {
	primitive tview.Primitive
	pages     *tview.Pages
	name      string
	mu        sync.Mutex
}

// Primitive creates an indicator that shows/hides a tview primitive.
// The primitive is added to pages when shown and removed when hidden.
//
// Example:
//
//	spinner := NewSpinner()
//	async.NewLoader[Data]().
//	    WithIndicator(async.Primitive(spinner, pages, "spinner")).
//	    Run(fetchData)
func Primitive(p tview.Primitive, pages *tview.Pages, name string) LoadingIndicator {
	return &primitiveIndicator{
		primitive: p,
		pages:     pages,
		name:      name,
	}
}

func (p *primitiveIndicator) Show() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pages != nil && p.primitive != nil {
		p.pages.AddPage(p.name, p.primitive, true, true)
	}
}

func (p *primitiveIndicator) Hide() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pages != nil {
		p.pages.RemovePage(p.name)
	}
}

func (p *primitiveIndicator) Success()        {}
func (p *primitiveIndicator) Error(err error) {}

// --- Noop Indicator ---

// noopIndicator does nothing (for testing or when no indicator is wanted).
type noopIndicator struct{}

// Noop returns an indicator that does nothing.
// Use this explicitly when you want no loading feedback.
func Noop() LoadingIndicator {
	return &noopIndicator{}
}

func (n *noopIndicator) Show()           {}
func (n *noopIndicator) Hide()           {}
func (n *noopIndicator) Success()        {}
func (n *noopIndicator) Error(err error) {}

// --- Multi Indicator ---

// multiIndicator combines multiple indicators.
type multiIndicator struct {
	indicators []LoadingIndicator
}

// Multi combines multiple indicators into one.
// All indicators are shown/hidden together.
//
// Example:
//
//	async.NewLoader[Data]().
//	    WithIndicator(async.Multi(
//	        async.Toast("Loading..."),
//	        async.StatusBar("Loading data...", setStatus),
//	    )).
//	    Run(fetchData)
func Multi(indicators ...LoadingIndicator) LoadingIndicator {
	return &multiIndicator{indicators: indicators}
}

func (m *multiIndicator) Show() {
	for _, i := range m.indicators {
		i.Show()
	}
}

func (m *multiIndicator) Hide() {
	for _, i := range m.indicators {
		i.Hide()
	}
}

func (m *multiIndicator) Success() {
	for _, i := range m.indicators {
		i.Success()
	}
}

func (m *multiIndicator) Error(err error) {
	for _, i := range m.indicators {
		i.Error(err)
	}
}
