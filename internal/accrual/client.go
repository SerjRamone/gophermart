// Package accrual ...
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/internal/server/handlers"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

var (
	// ErrUnknownOrder 'unknown order' error
	ErrUnknownOrder = errors.New("unknown order")

	// ErrTooManyRequests 'too many requests' error
	ErrTooManyRequests = errors.New("too many requests")
)

// AccrualClient ...
type AccrualClient struct {
	httpClient *http.Client
	url        string
}

// HTTPError ...
type HTTPError struct {
	Err        error
	StatusCode int
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%v (Code: %v)", e.Err, e.StatusCode)
}

func newHTTPError(e error, code int) error {
	return &HTTPError{
		Err:        e,
		StatusCode: code,
	}
}

// NewAccrualClient constructor
func NewAccrualClient(accrualURL string) *AccrualClient {
	return &AccrualClient{
		httpClient: &http.Client{},
		url:        accrualURL,
	}
}

// getAccrualData do request to accrual service, returns order accrual data
func (c AccrualClient) getAccrualData(ctx context.Context, order *models.Order) (*models.OrderAccrual, error) {
	// build url
	url, err := url.JoinPath(c.url, "/api/orders/", order.Number)
	if err != nil {
		return nil, fmt.Errorf("build url error: %w", err)
	}

	// make request object
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("accrual request creation error: %w", err)
	}

	// do http request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("accrual request doing error: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("response body close error", zap.Error(err))
		}
	}()

	// check statuses
	switch resp.StatusCode {
	case http.StatusTooManyRequests: // 429
		return nil, newHTTPError(errors.New("too many request status from accrual service"), resp.StatusCode)
	case http.StatusNoContent: // 204
		return nil, newHTTPError(errors.New("order is not registered"), resp.StatusCode)
	case http.StatusInternalServerError: // 500
		return nil, newHTTPError(errors.New("accrual service internal error"), resp.StatusCode)
	}

	// read response body
	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body error: %w", err)
	}

	// unmarshal response body
	var orderAccrual models.OrderAccrual
	if err := json.Unmarshal(rBody, &orderAccrual); err != nil {
		return nil, fmt.Errorf("unmarshal response body error: %w", err)
	}

	return &orderAccrual, nil
}

// WatchOrders starts order processing
func (c AccrualClient) WatchOrders(ctx context.Context, db handlers.Storage) {
	logger.Info("accrual client: watch orders process started")

	ticker := time.NewTicker(2 * time.Second)
	delayCh := make(chan int, 1)
	errCh := make(chan error, 1)
	ordersCh := make(chan *models.Order, 5) // @todo to config?

	// gets new orders from storage and puts to chan
	go func(ctx context.Context, ordersCh chan<- *models.Order, delayCh <-chan int, errCh chan<- error) {
		for {
			select {
			case <-ctx.Done():
				return
			case timeout := <-delayCh:
				time.Sleep(time.Duration(timeout) * time.Second)
			case <-ticker.C:
			}

			// get unprocessed orders
			orders, err := db.GetUnprocessedOrders(ctx)
			if err != nil {
				errCh <- fmt.Errorf("get unprocessed orders error: %w", err)
				continue
			}
			logger.Info("unprocessed orders found", zap.Int("count", len(orders)))

			// put orders to chan
			for _, order := range orders {
				ordersCh <- order
			}
		}
	}(ctx, ordersCh, delayCh, errCh)

	// process orders from chan
	go func(ctx context.Context, ordersCh <-chan *models.Order, delayCh chan<- int, errCh chan<- error) {
		// process orders from chan
		for order := range ordersCh {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// get order data from accrual service
			orderAcc, err := c.getAccrualData(ctx, order)
			if err != nil {
				if errors.Is(err, ErrUnknownOrder) {
					continue
				}
				if errors.Is(err, ErrTooManyRequests) {
					// some cooldown
					timeout, ok := 10, true // @todo
					if ok {
						delayCh <- timeout
						time.Sleep(time.Duration(timeout) * time.Second)
						continue
					}
				}
				errCh <- fmt.Errorf("get accrual data error: %w", err)
				continue
			}

			// set new status and accrual points
			order.Status = orderAcc.Status
			order.Accrual = orderAcc.Accrual

			// store update
			if err := db.UpdateOrder(ctx, order); err != nil {
				errCh <- fmt.Errorf("update order error: %w", err)
			}
		}
	}(ctx, ordersCh, delayCh, errCh)

	// errors chan watcher
	// log errors from gorutines
	go func(ctx context.Context, errCh <-chan error) {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errCh:
				logger.Error("run order error", zap.Error(err))
			}
		}
	}(ctx, errCh)
}
