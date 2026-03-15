package measure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type MeasureResult struct {
	FirstLossAt     *time.Time
	FirstRecoveryAt *time.Time
	TotalSamples    int
	PacketsLost     int
}

type Measurer struct {
	target string
	cancel context.CancelFunc
	wg     sync.WaitGroup
	result MeasureResult
}

func NewMeasurer(target string) *Measurer {
	return &Measurer{target: target}
}

func (m *Measurer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	m.wg.Add(1)
	go m.run(ctx)
}

func (m *Measurer) Stop() MeasureResult {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
	return m.result
}

func (m *Measurer) run(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var (
		totalSamples    int
		packetsLost     int
		firstLossAt     *time.Time
		firstRecoveryAt *time.Time

		// state tracking
		hadSuccess    bool // at least one successful probe has been seen
		inLoss        bool // currently in a loss streak
		earlyFailures int  // count of initial failures before any success
		warnedEarly   bool // early-unreachable warning already emitted
	)

	for {
		select {
		case <-ctx.Done():
			m.result = MeasureResult{
				FirstLossAt:     firstLossAt,
				FirstRecoveryAt: firstRecoveryAt,
				TotalSamples:    totalSamples,
				PacketsLost:     packetsLost,
			}
			return

		case <-ticker.C:
			sendTime := time.Now()
			success := m.ping()

			totalSamples++

			if success {
				if !hadSuccess {
					hadSuccess = true
				}
				if inLoss {
					// first recovery after a loss streak
					t := sendTime
					firstRecoveryAt = &t
					inLoss = false
				}
			} else {
				packetsLost++

				if !hadSuccess {
					earlyFailures++
					if earlyFailures >= 3 && !warnedEarly {
						fmt.Fprintf(os.Stderr,
							"WARNING: target %s unreachable before fault — measurements may be inaccurate\n",
							m.target)
						warnedEarly = true
					}
				} else if !inLoss {
					// first loss after at least one success
					t := sendTime
					firstLossAt = &t
					inLoss = true
				}
			}
		}
	}
}

func (m *Measurer) ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", m.target)
	err := cmd.Run()
	return err == nil
}
