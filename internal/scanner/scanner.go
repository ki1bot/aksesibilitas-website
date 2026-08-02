package scanner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/ki1bot/aksesibilitas-website/embedded"
	"github.com/ki1bot/aksesibilitas-website/internal/security"
)

type Scanner struct {
	chromePath string
}

type Result struct {
	URL        string      `json:"url"`
	Title      string      `json:"title"`
	Language   string      `json:"language"`
	DurationMS int64       `json:"duration_ms"`
	Violations []Violation `json:"violations"`
}

type Violation struct {
	ID          string   `json:"id"`
	Impact      string   `json:"impact"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Help        string   `json:"help"`
	HelpURL     string   `json:"helpUrl"`
	Nodes       []Node   `json:"nodes"`
}

type Node struct {
	HTML           string   `json:"html"`
	Target         []string `json:"target"`
	FailureSummary string   `json:"failureSummary"`
}

type axeResponse struct {
	Violations []Violation `json:"violations"`
	Error      string      `json:"error"`
}

func New(chromePath string) *Scanner {
	return &Scanner{
		chromePath: strings.TrimSpace(chromePath),
	}
}

func (scanner *Scanner) Scan(
	ctx context.Context,
	rawURL string,
) (Result, error) {
	if !strings.Contains(
		embedded.AxeSource,
		"axe.version",
	) {
		return Result{}, errors.New(
			"axe-core belum disinkronkan; jalankan pnpm sync:axe",
		)
	}

	normalizedURL, err := security.ValidatePublicHTTPURL(
		ctx,
		rawURL,
	)
	if err != nil {
		return Result{}, err
	}

	profileDir, err := os.MkdirTemp(
		"",
		"aksescheck-chromium-",
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"gagal membuat profile Chromium sementara: %w",
			err,
		)
	}
	defer os.RemoveAll(profileDir)

	options := append(
		[]chromedp.ExecAllocatorOption{},
		chromedp.DefaultExecAllocatorOptions[:]...,
	)

	options = append(
		options,
		chromedp.UserDataDir(profileDir),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)

	if scanner.chromePath != "" {
		options = append(
			options,
			chromedp.ExecPath(scanner.chromePath),
		)
	}

	allocatorContext, cancelAllocator :=
		chromedp.NewExecAllocator(ctx, options...)
	defer cancelAllocator()

	browserContext, cancelBrowser :=
		chromedp.NewContext(allocatorContext)
	defer cancelBrowser()

	var requestCount atomic.Int32
	var documentCount atomic.Int32
	var receivedBytes atomic.Int64
	var blockedReason string
	var blockedMutex sync.Mutex

	setBlockedReason := func(reason string) {
		blockedMutex.Lock()
		defer blockedMutex.Unlock()

		if blockedReason == "" {
			blockedReason = reason
		}
	}

	getBlockedReason := func() string {
		blockedMutex.Lock()
		defer blockedMutex.Unlock()

		return blockedReason
	}

	chromedp.ListenTarget(
		browserContext,
		func(event any) {
			if data, ok :=
				event.(*network.EventDataReceived); ok {
				if receivedBytes.Add(
					int64(data.EncodedDataLength),
				) > 25<<20 {
					setBlockedReason(
						"total data halaman melebihi batas 25 MB",
					)
					cancelBrowser()
				}

				return
			}

			paused, ok :=
				event.(*fetch.EventRequestPaused)
			if !ok {
				return
			}

			go func() {
				contextData :=
					chromedp.FromContext(browserContext)

				if contextData == nil ||
					contextData.Target == nil {
					return
				}

				executorContext := cdp.WithExecutor(
					browserContext,
					contextData.Target,
				)

				requestedURL := paused.Request.URL
				allow := requestedURL == "about:blank"

				if strings.HasPrefix(
					requestedURL,
					"http://",
				) || strings.HasPrefix(
					requestedURL,
					"https://",
				) {
					if requestCount.Add(1) > 250 {
						setBlockedReason(
							"jumlah request halaman melebihi batas 250",
						)

						_ = fetch.FailRequest(
							paused.RequestID,
							network.ErrorReasonBlockedByClient,
						).Do(executorContext)

						return
					}

					if paused.ResourceType ==
						network.ResourceTypeDocument &&
						documentCount.Add(1) > 6 {
						setBlockedReason(
							"jumlah navigasi atau redirect melebihi batas lima",
						)

						_ = fetch.FailRequest(
							paused.RequestID,
							network.ErrorReasonBlockedByClient,
						).Do(executorContext)

						return
					}

					validationContext, cancel :=
						context.WithTimeout(
							browserContext,
							5*time.Second,
						)

					_, validationErr :=
						security.ValidatePublicHTTPURL(
							validationContext,
							requestedURL,
						)

					cancel()

					if validationErr == nil {
						allow = true
					} else {
						setBlockedReason(
							validationErr.Error(),
						)
					}
				}

				if !allow {
					_ = fetch.FailRequest(
						paused.RequestID,
						network.ErrorReasonBlockedByClient,
					).Do(executorContext)

					return
				}

				_ = fetch.ContinueRequest(
					paused.RequestID,
				).Do(executorContext)
			}()
		},
	)

	startedAt := time.Now()

	var pageTitle string
	var pageLanguage string
	var finalURL string
	var axeResult axeResponse

	err = chromedp.Run(
		browserContext,
		network.Enable(),
		browser.SetDownloadBehavior(
			browser.SetDownloadBehaviorBehaviorDeny,
		),
		fetch.Enable().WithPatterns(
			[]*fetch.RequestPattern{
				{
					URLPattern:   "*",
					RequestStage: fetch.RequestStageRequest,
				},
			},
		),
		chromedp.Navigate(normalizedURL),
		chromedp.WaitReady(
			"body",
			chromedp.ByQuery,
		),
		chromedp.Evaluate(
			`document.title || ""`,
			&pageTitle,
		),
		chromedp.Evaluate(
			`document.documentElement.lang || ""`,
			&pageLanguage,
		),
		chromedp.Evaluate(
			`location.href`,
			&finalURL,
		),
		chromedp.Evaluate(
			embedded.AxeSource,
			nil,
		),
		chromedp.Evaluate(
			`new Promise((resolve) => {
				axe.run(document, {
					runOnly: {
						type: "tag",
						values: [
							"wcag2a",
							"wcag2aa",
							"wcag21a",
							"wcag21aa",
							"wcag22a",
							"wcag22aa"
						]
					}
				}, (error, results) => {
					if (error) {
						resolve({
							error: error.message,
							violations: []
						});
						return;
					}

					resolve({
						error: "",
						violations: results.violations
					});
				});
			})`,
			&axeResult,
			func(
				params *runtime.EvaluateParams,
			) *runtime.EvaluateParams {
				return params.WithAwaitPromise(true)
			},
		),
	)

	if err != nil {
		if reason := getBlockedReason(); reason != "" {
			return Result{}, errors.New(reason)
		}

		return Result{}, fmt.Errorf(
			"pemindaian Chromium gagal: %w",
			err,
		)
	}

	if axeResult.Error != "" {
		return Result{}, errors.New(
			axeResult.Error,
		)
	}

	return Result{
		URL:        finalURL,
		Title:      pageTitle,
		Language:   pageLanguage,
		DurationMS: time.Since(startedAt).Milliseconds(),
		Violations: axeResult.Violations,
	}, nil
}
