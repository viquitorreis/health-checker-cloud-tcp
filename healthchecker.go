package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func TCPHCTypeValidator(t string) bool {
	return validHCTypes[t]
}

func NewTCPOpts(hcType string, ip net.IP, port, packets int, url string, isAuth bool) *TCPOptions {
	if !TCPHCTypeValidator(hcType) {
		log.Fatalf("Health checker type not allowed: %s", hcType)
	}

	switch hcType {
	case "tcp":
		if ip == nil || port == 0 {
			log.Fatalf("For 'tcp', IP and Port must be specified")
		}
	case "cloud":
		if url == "" {
			log.Fatalf("For 'cloud', URL must be specified")
		}
	}

	return &TCPOptions{
		hcType:  TCPHCType{typ: hcType},
		ip:      ip,
		port:    port,
		packets: packets,
		URL:     url,
		isAuth:  isAuth,
	}
}

func NewTCPChecker(opts *TCPOptions) *TCPChecker {
	if opts == nil {
		log.Fatal("TCPOptions cannot be nil")
	}

	// Ensure only relevant fields are set
	switch opts.hcType.typ {
	case "tcp":
		if opts.ip == nil || opts.port == 0 {
			log.Fatal("TCP health-check requires valid IP and Port")
		}
	case "cloud":
		if opts.URL == "" {
			log.Fatal("Cloud health-check requires a valid URL")
		}
	}

	return &TCPChecker{
		Target: Target{
			IP:      opts.ip,
			Port:    opts.port,
			Packets: opts.packets,
			URL:     opts.URL,
		},
		TCPOptions: *opts,
	}
}

func (hc *TCPChecker) addr() string {
	// Only for TCP
	if hc.TCPOptions.hcType.typ != "tcp" {
		log.Fatal("addr() called for non-TCP health-check type")
	}
	return fmt.Sprintf("%s:%d", hc.IP.String(), hc.Port)
}

func (hc *TCPChecker) DoAuth() (int, error) {
	if hc.Target.URL == "" {
		return 0, fmt.Errorf("No target provided")
	}

	username := os.Getenv("USER_USERNAME")
	pw := os.Getenv("USER_PASSWORD")

	if username == "" || pw == "" {
		return 0, fmt.Errorf("Missing username or password in environment variables")
	}

	user := userAuth{
		Username: username,
		Password: pw,
	}

	body, _ := json.Marshal(user)

	payload := bytes.NewBuffer(body)

	resp, err := http.Post(hc.Target.URL, "application/json", payload)
	if err != nil {
		fmt.Printf("Error while doing post request: %v", err)
		return 0, fmt.Errorf("Error while doing post request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func sendAlertEmail(to []string, subject, body string) error {
	emailReq := NewEmailRequest(to, subject, body)
	templateData := struct{ Message string }{Message: body}
	if err := emailReq.ParseTemplate("hc_alert.html", templateData); err != nil {
		return fmt.Errorf("error parsing email template: %w", err)
	}
	if err := emailReq.SendEmail(); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}

func (hc *TCPChecker) Check(timeout time.Duration) *Result {
	switch hc.TCPOptions.hcType.typ {
	case "tcp":
		return hc.performTCPCheck(timeout)
	case "cloud":
		return hc.performCloudCheck()
	default:
		return &Result{Success: false, Message: "Unknown health check type"}
	}
}

func (hc *TCPChecker) performTCPCheck(timeout time.Duration) *Result {
	conn, err := net.DialTimeout("tcp", hc.addr(), timeout)
	if err != nil {
		return &Result{Success: false, Message: fmt.Sprintf("Failed to connect: %v", err)}
	}
	defer conn.Close()
	return &Result{Success: true, Message: "TCP connection successful"}
}

func (hc *TCPChecker) performCloudCheck() *Result {
	if !hc.TCPOptions.isAuth {
		return &Result{Success: true, Message: "Cloud health check without auth passed"}
	}
	status, err := hc.DoAuth()
	if err != nil || status != 200 {
		msg := fmt.Sprintf("Status code: %d\n Cloud check failed: %v\n", status, err)
		_ = sendAlertEmail([]string{
			"victorreisprog@gmail.com",
			"victorreis@biofy.tech",
			"guilhermejesus@biofy.tech",
			"pedroleao@biofy.tech",
		}, "⚠️ ALERT - API IS DOWN ⚠️", msg)
		return &Result{Success: false, Message: msg}
	}

	return &Result{Success: true, Message: "Cloud check passed"}
}

func (hc *TCPChecker) CheckWithRetries(retries int, retryDelay time.Duration, logOutput io.Writer) *Result {
	var result *Result
	attempt := 0

	for {
		start := time.Now()
		result = hc.Check(hc.Timeout)
		duration := time.Since(start)

		attempt++
		logOutput.Write([]byte(fmt.Sprintf("Health Check Attempt %d - Success: %v, Latency: %v, MSG: %s\n", attempt, result.Success, duration, result.Message)))

		// if result.Success {
		// 	return result
		// }

		if retries != -1 && attempt >= retries {
			break
		}

		time.Sleep(retryDelay)
	}

	return result
}
