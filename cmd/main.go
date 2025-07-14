/*
	upsmon - Minimal UPS monitoring utility. It shutdowns PC when UPS is offline
	for the specified time.

	Copyright (C) 2025 Vadim Kuznetsov <vimusov@gmail.com>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/sstallion/go-hid"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type config struct {
	debug   bool
	trace   bool
	refresh time.Duration
	delay   time.Duration
	vendor  uint
	product uint
}

func showErr(msg string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s, '%v'.\n", msg, err)
}

func showInfo(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "INFO: "+format+".\n", args...)
}

func parseFlags(state string) (bool, error) {
	fields := strings.Fields(state)
	if len(fields) < 1 {
		return false, errors.New("wrong columns count")
	}
	flags, err := strconv.ParseUint(fields[len(fields)-1], 2, 8)
	if err != nil {
		return false, err
	}
	return flags&0b10000000 != 0, nil
}

var queryState = func(cfg *config) (string, error) {
	dev, openErr := hid.OpenFirst(uint16(cfg.vendor), uint16(cfg.product))
	if openErr != nil {
		return "", openErr
	}
	defer func() { _ = dev.Close() }()

	if _, writeErr := dev.Write([]byte("Q1\r")); writeErr != nil {
		return "", writeErr
	}
	state, readErr := bufio.NewReader(dev).ReadString('\r')
	if readErr != nil {
		return "", readErr
	}
	return strings.TrimSpace(state), nil
}

var runShutdown = func() {
	cmd := exec.Command("systemctl", "poweroff")
	if stdout, execErr := cmd.CombinedOutput(); execErr != nil {
		showErr(fmt.Sprintf("Poweroff failed with '%s'", stdout), execErr)
	}
}

func checkStatus(cfg *config, deadline *time.Time) {
	state, queryErr := queryState(cfg)
	if queryErr != nil {
		showErr("Failed to query state", queryErr)
		return
	}
	if cfg.trace {
		showInfo("Raw state: '%s'", state)
	}

	isOffline, parseErr := parseFlags(state)
	if parseErr != nil {
		showErr("Failed to parse flags", parseErr)
		return
	}

	if isOffline && deadline.IsZero() {
		*deadline = time.Now().Add(cfg.delay)
		showInfo("UPS is offline, shutdown at %s", deadline.Format(time.DateTime))
		return
	}

	if !isOffline && !deadline.IsZero() {
		showInfo("UPS was offline and return to online")
		*deadline = time.Time{}
		return
	}

	if !deadline.IsZero() && time.Now().After(*deadline) {
		showInfo("Timeout is over, shutting down")
		runShutdown()
	}

	if cfg.debug {
		if isOffline {
			showInfo("UPS is offline")
		} else {
			showInfo("UPS is online")
		}
	}
}

func main() {
	var cfg config

	flag.BoolVar(&cfg.debug, "debug", false, "enable debug")
	flag.BoolVar(&cfg.trace, "trace", false, "enable trace")
	flag.DurationVar(&cfg.refresh, "refresh", 10*time.Second, "refresh interval")
	flag.DurationVar(&cfg.delay, "delay", 2*time.Minute, "delay before shutdown")
	flag.UintVar(&cfg.vendor, "vendor", 0x0665, "USB vendor")
	flag.UintVar(&cfg.product, "product", 0x5161, "USB product")
	flag.Parse()

	if initHidErr := hid.Init(); initHidErr != nil {
		showErr("Unable to init HID", initHidErr)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	deadline := time.Time{}
processing:
	for {
		checkStatus(&cfg, &deadline)
		select {
		case <-signals:
			break processing
		case <-time.After(cfg.refresh):
		}
	}

	_ = hid.Exit()
}
