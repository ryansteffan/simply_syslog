package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ryansteffan/simply_syslog/internal/config"
)

type runMode int

const (
	modeUDP runMode = iota
	modeTCP
	modeBoth
)

type options struct {
	clients    int
	duration   time.Duration
	mode       runMode
	host       string
	udpPort    string
	tcpPort    string
	baseRPS    int
	amplitude  float64
	period     time.Duration
	useConfig  bool
	configPath string
	seed       int64
}

func parseFlags() options {
	var (
		clients    = flag.Int("clients", 50, "number of concurrent simulated clients")
		duration   = flag.Duration("duration", time.Minute, "how long to run, e.g. 90s, 2m")
		protocol   = flag.String("protocol", "udp", "protocol to use: udp|tcp|both")
		host       = flag.String("host", "127.0.0.1", "target host for the syslog server")
		udpPort    = flag.String("udp-port", "514", "UDP port for the syslog server")
		tcpPort    = flag.String("tcp-port", "514", "TCP port for the syslog server")
		baseRPS    = flag.Int("base-rps", 1000, "approx total messages per second across all clients")
		amplitude  = flag.Float64("amplitude", 0.8, "load oscillation amplitude (0..1)")
		period     = flag.Duration("period", 30*time.Second, "oscillation period for peaks and lows")
		useConfig  = flag.Bool("use-config", false, "load address/ports from ./config/config.json (overrides host/ports)")
		configPath = flag.String("config", "./config/config.json", "path to server config.json when --use-config is set")
		seed       = flag.Int64("seed", 0, "random seed (0=use time)")
	)
	flag.Parse()

	m := modeUDP
	switch strings.ToLower(*protocol) {
	case "udp":
		m = modeUDP
	case "tcp":
		m = modeTCP
	case "both":
		m = modeBoth
	default:
		fmt.Println("Invalid --protocol; use udp|tcp|both")
		os.Exit(2)
	}

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	return options{
		clients:    *clients,
		duration:   *duration,
		mode:       m,
		host:       *host,
		udpPort:    *udpPort,
		tcpPort:    *tcpPort,
		baseRPS:    *baseRPS,
		amplitude:  *amplitude,
		period:     *period,
		useConfig:  *useConfig,
		configPath: *configPath,
		seed:       *seed,
	}
}

func loadTargetFromConfig(opts *options) {
	conf, err := config.LoadConfig(opts.configPath)
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(2)
	}
	// Server binds to 0.0.0.0; use localhost by default for client.
	opts.host = "127.0.0.1"
	if conf.Data.Udp_Port != "" {
		opts.udpPort = conf.Data.Udp_Port
	}
	if conf.Data.Tcp_Port != "" {
		opts.tcpPort = conf.Data.Tcp_Port
	}
	switch strings.ToUpper(conf.Data.Protocol) {
	case "UDP":
		opts.mode = modeUDP
	case "TCP":
		opts.mode = modeTCP
	case "BOTH":
		opts.mode = modeBoth
	}
}

func main() {
	opts := parseFlags()
	if opts.useConfig {
		loadTargetFromConfig(&opts)
	}

	rand.Seed(opts.seed)

	ctx, cancel := context.WithTimeout(context.Background(), opts.duration)
	defer cancel()

	var totalSent uint64
	var wg sync.WaitGroup

	// Precompute per-client base rate and random phase
	perClientBase := math.Max(1, float64(opts.baseRPS)/float64(opts.clients))

	start := time.Now()

	switch opts.mode {
	case modeUDP:
		udpAddr := net.JoinHostPort(opts.host, opts.udpPort)
		addr, err := net.ResolveUDPAddr("udp", udpAddr)
		if err != nil {
			fmt.Println("resolve UDP:", err)
			os.Exit(1)
		}
		for i := 0; i < opts.clients; i++ {
			phase := rand.Float64() * 2 * math.Pi
			wg.Add(1)
			go func(id int, ph float64) {
				defer wg.Done()
				udpClient(ctx, id, addr, perClientBase, opts.amplitude, opts.period, ph, &totalSent, start)
			}(i, phase)
		}
	case modeTCP, modeBoth:
		// Start TCP clients
		if opts.mode == modeTCP || opts.mode == modeBoth {
			tcpAddr := net.JoinHostPort(opts.host, opts.tcpPort)
			for i := 0; i < opts.clients; i++ {
				phase := rand.Float64() * 2 * math.Pi
				wg.Add(1)
				go func(id int, ph float64) {
					defer wg.Done()
					tcpClient(ctx, id, tcpAddr, perClientBase, opts.amplitude, opts.period, ph, &totalSent, start)
				}(i, phase)
			}
		}
		// If BOTH, also start UDP
		if opts.mode == modeBoth {
			udpAddr := net.JoinHostPort(opts.host, opts.udpPort)
			addr, err := net.ResolveUDPAddr("udp", udpAddr)
			if err != nil {
				fmt.Println("resolve UDP:", err)
				os.Exit(1)
			}
			for i := 0; i < opts.clients; i++ {
				phase := rand.Float64() * 2 * math.Pi
				wg.Add(1)
				go func(id int, ph float64) {
					defer wg.Done()
					udpClient(ctx, id, addr, perClientBase, opts.amplitude, opts.period, ph, &totalSent, start)
				}(i, phase)
			}
		}
	}

	// Progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sent := atomic.LoadUint64(&totalSent)
				elapsed := time.Since(start).Seconds()
				rate := float64(sent) / math.Max(elapsed, 0.001)
				fmt.Printf("[bench] sent=%d elapsed=%.1fs avg_rps=%.1f\n", sent, elapsed, rate)
			}
		}
	}()

	wg.Wait()
	sent := atomic.LoadUint64(&totalSent)
	elapsed := time.Since(start).Seconds()
	rate := float64(sent) / math.Max(elapsed, 0.001)
	fmt.Printf("Done. sent=%d elapsed=%.1fs avg_rps=%.1f\n", sent, elapsed, rate)
}

func udpClient(
	ctx context.Context,
	id int,
	addr *net.UDPAddr,
	base float64,
	amp float64,
	period time.Duration,
	phase float64,
	total *uint64,
	start time.Time,
) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("udp dial:", err)
		return
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg := genRFC3164Message(id)
		// UDP syslog typically doesn't require newline
		_, _ = conn.Write(msg)
		atomic.AddUint64(total, 1)

		// dynamic pacing
		dt := nextDelay(base, amp, period, phase, start)
		time.Sleep(dt)
	}
}

func tcpClient(
	ctx context.Context,
	id int,
	address string,
	base float64,
	amp float64,
	period time.Duration,
	phase float64,
	total *uint64,
	start time.Time,
) {
	// reconnect loop for long runs
	var conn net.Conn
	var err error
	dial := func() bool {
		conn, err = net.DialTimeout("tcp", address, 3*time.Second)
		if err != nil {
			// backoff and retry until context done
			select {
			case <-ctx.Done():
				return false
			case <-time.After(2 * time.Second):
			}
			return true
		}
		return true
	}
	if !dial() {
		return
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg := genRFC3164Message(id)
		// Most TCP syslog receivers expect LF-delimited messages
		_, err = conn.Write(append(msg, '\n'))
		if err != nil {
			// attempt to reconnect
			if conn != nil {
				conn.Close()
			}
			if !dial() {
				return
			}
			continue
		}
		atomic.AddUint64(total, 1)

		dt := nextDelay(base, amp, period, phase, start)
		time.Sleep(dt)
	}
}

func nextDelay(base, amp float64, period time.Duration, phase float64, start time.Time) time.Duration {
	if base < 1 {
		base = 1
	}
	t := time.Since(start).Seconds()
	per := math.Max(period.Seconds(), 1)
	mod := 1 + amp*math.Sin((2*math.Pi*t)/per+phase)
	// occasional burst
	if rand.Float64() < 0.02 { // 2% chance to enter a brief burst
		mod *= 2 + rand.Float64()*1.5 // 2x to 3.5x
	}
	rate := base * mod
	// jitter  +/-20%
	jitter := 0.8 + rand.Float64()*0.4
	if rate < 1 {
		rate = 1
	}
	d := time.Second / time.Duration(rate)
	return time.Duration(float64(d) * jitter)
}

func genRFC3164Message(clientID int) []byte {
	// <pri>timestamp hostname tag[pid]: message
	pri := 14 // default user.info
	ts := time.Now().Format("Jan 02 15:04:05")
	host := fmt.Sprintf("host-%02d", clientID%99)
	tag := fmt.Sprintf("app%d", clientID%10)
	pid := 1000 + rand.Intn(9000)
	// variable message size
	base := fmt.Sprintf("Synthetic load from client=%d", clientID)
	extraLen := 10 + rand.Intn(120)
	extra := randString(extraLen)
	msg := fmt.Sprintf("<%d>%s %s %s[%d]: %s %s", pri, ts, host, tag, pid, base, extra)
	return []byte(msg)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
