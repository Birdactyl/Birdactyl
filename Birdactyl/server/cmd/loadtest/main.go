package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type flags struct {
	URL      string
	Users    int
	Duration time.Duration
	Prefix   string
	Workers  int
}

type userCtx struct {
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
	SpoofedIP    string
}

type result struct {
	Endpoint   string
	StatusCode int
	Latency    time.Duration
	IsError    bool
}

type endpointStats struct {
	Requests  int64
	Errors    int64
	Latencies []time.Duration
	StatusMap map[int]int64
}

type weightedEndpoint struct {
	Weight   int
	Method   string
	Path     string
	NeedAuth bool
}

var endpoints = []weightedEndpoint{
	{Weight: 35, Method: "GET", Path: "/api/v1/health", NeedAuth: false},
	{Weight: 25, Method: "GET", Path: "/api/v1/auth/me", NeedAuth: true},
	{Weight: 20, Method: "GET", Path: "/api/v1/servers/", NeedAuth: true},
	{Weight: 10, Method: "GET", Path: "/api/v1/auth/sessions", NeedAuth: true},
	{Weight: 10, Method: "GET", Path: "/api/v1/auth/resources", NeedAuth: true},
}

func main() {
	f := parseFlags()

	fmt.Println("=== Birdactyl Load Test ===")
	fmt.Printf("Target:     %s\n", f.URL)
	fmt.Printf("Users:      %d\n", f.Users)
	fmt.Printf("Duration:   %s\n", f.Duration)
	fmt.Println()

	if !checkHealth(f.URL) {
		fmt.Fprintf(os.Stderr, "Server at %s is not reachable. Make sure it's running.\n", f.URL)
		os.Exit(1)
	}

	fmt.Printf("[Setup] Registering %d test users... ", f.Users)
	start := time.Now()
	users := registerUsersParallel(f)
	if len(users) == 0 {
		fmt.Fprintf(os.Stderr, "\nFailed to register any users. Check server logs.\n")
		os.Exit(1)
	}
	fmt.Printf("done (%d/%d in %s)\n", len(users), f.Users, time.Since(start).Round(time.Millisecond))

	fmt.Printf("[Setup] Logging in users without tokens... ")
	start = time.Now()
	loginUsersParallel(f, users)
	authedCount := 0
	for _, u := range users {
		if u.AccessToken != "" {
			authedCount++
		}
	}
	fmt.Printf("done (%d/%d in %s)\n", authedCount, len(users), time.Since(start).Round(time.Millisecond))
	if authedCount == 0 {
		fmt.Fprintf(os.Stderr, "\nNo users could authenticate. Check server logs.\n")
		os.Exit(1)
	}

	fmt.Printf("\n[Load] Running for %s with %d concurrent users...\n", f.Duration, len(users))

	results := runLoadTest(f, users)
	printReport(results, f.Duration)
}

func parseFlags() flags {
	f := flags{}
	flag.StringVar(&f.URL, "url", "http://localhost:3000", "Base URL of the panel server")
	flag.IntVar(&f.Users, "users", 50, "Number of concurrent virtual users")
	flag.DurationVar(&f.Duration, "duration", 30*time.Second, "Duration of the load test")
	flag.StringVar(&f.Prefix, "prefix", "loadtest", "Prefix for test user accounts")
	flag.IntVar(&f.Workers, "workers", 50, "Concurrent workers for setup phase")
	flag.Parse()
	f.URL = strings.TrimRight(f.URL, "/")
	if f.Workers > f.Users {
		f.Workers = f.Users
	}
	return f
}

func newHTTPClient(maxConns int) *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        maxConns,
			MaxIdleConnsPerHost: maxConns,
			MaxConnsPerHost:     maxConns,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

func checkHealth(baseURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/v1/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func registerUsersParallel(f flags) []*userCtx {
	client := newHTTPClient(f.Workers)
	users := make([]*userCtx, f.Users)
	for i := 0; i < f.Users; i++ {
		users[i] = &userCtx{
			Email:     fmt.Sprintf("%s_%d@loadtest.local", f.Prefix, i),
			Password:  fmt.Sprintf("%s_pass_%d_secure!", f.Prefix, i),
			SpoofedIP: fmt.Sprintf("10.%d.%d.%d", (i/65536)%256, (i/256)%256, i%256),
		}
	}

	sem := make(chan struct{}, f.Workers)
	var wg sync.WaitGroup

	for _, u := range users {
		wg.Add(1)
		sem <- struct{}{}
		go func(u *userCtx) {
			defer wg.Done()
			defer func() { <-sem }()

			body, _ := json.Marshal(map[string]string{
				"email":    u.Email,
				"username": strings.TrimSuffix(u.Email, "@loadtest.local"),
				"password": u.Password,
			})

			req, err := http.NewRequest("POST", f.URL+"/api/v1/auth/register", bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", u.SpoofedIP)
			resp, err := client.Do(req)
			if err != nil {
				return
			}

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()

			if resp.StatusCode == 201 {
				if data, ok := result["data"].(map[string]interface{}); ok {
					if tokens, ok := data["tokens"].(map[string]interface{}); ok {
						u.AccessToken, _ = tokens["access_token"].(string)
						u.RefreshToken, _ = tokens["refresh_token"].(string)
					}
				}
			}
		}(u)
	}
	wg.Wait()

	return users
}

func loginUsersParallel(f flags, users []*userCtx) {
	needLogin := make([]*userCtx, 0)
	for _, u := range users {
		if u.AccessToken == "" {
			needLogin = append(needLogin, u)
		}
	}
	if len(needLogin) == 0 {
		return
	}

	client := newHTTPClient(f.Workers)
	sem := make(chan struct{}, f.Workers)
	var wg sync.WaitGroup

	for _, u := range needLogin {
		wg.Add(1)
		sem <- struct{}{}
		go func(u *userCtx) {
			defer wg.Done()
			defer func() { <-sem }()

			body, _ := json.Marshal(map[string]string{
				"email":    u.Email,
				"password": u.Password,
			})

			req, err := http.NewRequest("POST", f.URL+"/api/v1/auth/login", bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Forwarded-For", u.SpoofedIP)
			resp, err := client.Do(req)
			if err != nil {
				return
			}

			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()

			if resp.StatusCode == 200 {
				if data, ok := result["data"].(map[string]interface{}); ok {
					if tokens, ok := data["tokens"].(map[string]interface{}); ok {
						u.AccessToken, _ = tokens["access_token"].(string)
						u.RefreshToken, _ = tokens["refresh_token"].(string)
					}
				}
			}
		}(u)
	}
	wg.Wait()
}

func runLoadTest(f flags, users []*userCtx) []result {
	resultsCh := make(chan result, 10000)
	var done int32
	var wg sync.WaitGroup

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	deadline := time.After(f.Duration)
	client := newHTTPClient(f.Users)

	totalWeight := 0
	for _, ep := range endpoints {
		totalWeight += ep.Weight
	}

	for i := 0; i < len(users); i++ {
		wg.Add(1)
		go func(user *userCtx, workerID int) {
			defer wg.Done()
			requestNum := 0

			for atomic.LoadInt32(&done) == 0 {
				pick := requestNum % totalWeight
				var ep weightedEndpoint
				cumulative := 0
				for _, candidate := range endpoints {
					cumulative += candidate.Weight
					if pick < cumulative {
						ep = candidate
						break
					}
				}

				if ep.NeedAuth && user.AccessToken == "" {
					ep = endpoints[0]
				}

				r := fireRequest(client, f.URL, ep, user)
				resultsCh <- r
				requestNum++
			}
		}(users[i], i)
	}

	var results []result
	var collectorDone sync.WaitGroup
	collectorDone.Add(1)
	go func() {
		defer collectorDone.Done()
		for r := range resultsCh {
			results = append(results, r)
		}
	}()

	select {
	case <-deadline:
	case <-stop:
		fmt.Println("\nInterrupted.")
	}
	atomic.StoreInt32(&done, 1)
	wg.Wait()
	close(resultsCh)
	collectorDone.Wait()

	return results
}

func fireRequest(client *http.Client, baseURL string, ep weightedEndpoint, user *userCtx) result {
	url := baseURL + ep.Path
	req, err := http.NewRequest(ep.Method, url, nil)
	if err != nil {
		return result{Endpoint: ep.Method + " " + ep.Path, IsError: true}
	}

	req.Header.Set("X-Forwarded-For", user.SpoofedIP)
	if ep.NeedAuth && user.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+user.AccessToken)
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return result{
			Endpoint: ep.Method + " " + ep.Path,
			Latency:  latency,
			IsError:  true,
		}
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	isErr := resp.StatusCode >= 500
	return result{
		Endpoint:   ep.Method + " " + ep.Path,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		IsError:    isErr,
	}
}

func printReport(results []result, duration time.Duration) {
	if len(results) == 0 {
		fmt.Println("\nNo results collected.")
		return
	}

	stats := make(map[string]*endpointStats)
	globalStatus := make(map[int]int64)
	var totalErrors int64

	for _, r := range results {
		ep, ok := stats[r.Endpoint]
		if !ok {
			ep = &endpointStats{StatusMap: make(map[int]int64)}
			stats[r.Endpoint] = ep
		}
		ep.Requests++
		ep.Latencies = append(ep.Latencies, r.Latency)
		if r.StatusCode > 0 {
			ep.StatusMap[r.StatusCode]++
			globalStatus[r.StatusCode]++
		}
		if r.IsError {
			ep.Errors++
			totalErrors++
		}
	}

	totalReqs := int64(len(results))
	throughput := float64(totalReqs) / duration.Seconds()
	errorRate := float64(totalErrors) / float64(totalReqs) * 100

	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Total Requests:  %s\n", fmtInt(totalReqs))
	fmt.Printf("Total Duration:  %s\n", duration.Round(time.Millisecond))
	fmt.Printf("Throughput:      %.1f req/s\n", throughput)
	fmt.Printf("Error Rate:      %.2f%%\n", errorRate)
	fmt.Println()

	nameWidth := 30
	fmt.Printf("%-*s  %8s  %6s  %8s  %8s  %8s  %8s  %8s\n",
		nameWidth, "Endpoint", "Requests", "Errors", "Avg", "P50", "P95", "P99", "Max")
	fmt.Println(strings.Repeat("-", nameWidth+2+8+2+6+2+8+2+8+2+8+2+8+2+8))

	names := make([]string, 0, len(stats))
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		ep := stats[name]
		sort.Slice(ep.Latencies, func(i, j int) bool { return ep.Latencies[i] < ep.Latencies[j] })

		avg := avgDuration(ep.Latencies)
		p50 := percentile(ep.Latencies, 50)
		p95 := percentile(ep.Latencies, 95)
		p99 := percentile(ep.Latencies, 99)
		maxL := ep.Latencies[len(ep.Latencies)-1]

		displayName := name
		if len(displayName) > nameWidth {
			displayName = displayName[:nameWidth-2] + ".."
		}

		fmt.Printf("%-*s  %8s  %6d  %8s  %8s  %8s  %8s  %8s\n",
			nameWidth, displayName,
			fmtInt(ep.Requests), ep.Errors,
			fmtDur(avg), fmtDur(p50), fmtDur(p95), fmtDur(p99), fmtDur(maxL))
	}

	fmt.Println()
	fmt.Print("Status Codes: ")
	codes := make([]int, 0, len(globalStatus))
	for code := range globalStatus {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, fmt.Sprintf("%d=%s", code, fmtInt(globalStatus[code])))
	}
	fmt.Println(strings.Join(parts, "  "))
	fmt.Println()
}

func avgDuration(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range ds {
		total += d
	}
	return total / time.Duration(len(ds))
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func fmtDur(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fus", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func fmtInt(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	rem := len(s) % 3
	if rem > 0 {
		b.WriteString(s[:rem])
	}
	for i := rem; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
