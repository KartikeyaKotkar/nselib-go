package nsdlfpi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"
)

// launchArgs mirrors Python's BROWSER_LAUNCH_ARGS.
var launchArgs = []string{
	"--no-sandbox",
	"--disable-setuid-sandbox",
	"--disable-dev-shm-usage",
}

// Browser drives headless Chrome for the NSDL production site.
// Equivalent to Python's NSDLProductionBrowser (pyppeteer).
type Browser struct {
	executable string
}

// NewBrowser resolves a Chrome/Edge executable, or "" when absent.
func NewBrowser() *Browser {
	return &Browser{executable: findBrowserExecutable()}
}

func findBrowserExecutable() string {
	if runtime.GOOS == "windows" {
		for _, c := range []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		} {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		return ""
	}
	for _, c := range []string{"google-chrome", "chromium", "chromium-browser", "chrome", "microsoft-edge"} {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	return ""
}

// Available reports whether a browser executable was found.
func (b *Browser) Available() bool { return b != nil && b.executable != "" }

// Close releases browser resources. Contexts are per-call and self-cleaning,
// so this is a no-op kept for parity with Python's NSDLProductionBrowser.close.
func (b *Browser) Close() {}

func (b *Browser) run(tasks ...chromedp.Action) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(b.executable),
		chromedp.Flag("headless", true),
	)
	for _, a := range launchArgs {
		name := a[2:]
		opts = append(opts, chromedp.Flag(name, true))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, 90*time.Second)
	defer cancelTimeout()
	return chromedp.Run(ctx, tasks...)
}

// LatestHTML renders Latest.aspx and returns page HTML.
func (b *Browser) LatestHTML() (string, error) {
	var html string
	err := b.run(
		chromedp.Navigate(ProductionBaseURL+"/"+LatestPage),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	return html, err
}

// ArchiveHTML fills the archive date form, submits, and returns result HTML.
func (b *Browser) ArchiveHTML(tradeDate time.Time) (string, error) {
	dateText := tradeDate.Format(ReportDateLayout)
	var html string
	err := b.run(
		chromedp.Navigate(ProductionBaseURL+"/"+ArchivePage),
		chromedp.WaitVisible("#txtDate", chromedp.ByID),
		chromedp.Evaluate(fmt.Sprintf(`(function(){
			const txtDate = document.querySelector('#txtDate');
			const hiddenDate = document.querySelector('#hdnDate');
			if (!txtDate || !hiddenDate) {
				throw new Error('NSDL archive date controls were not found');
			}
			txtDate.disabled = false;
			txtDate.value = %q;
			hiddenDate.value = %q;
		})()`, dateText, dateText), nil),
		chromedp.Click("#btnSubmit1", chromedp.ByID),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	return html, err
}
