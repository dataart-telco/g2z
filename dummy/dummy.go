/*
 * g2z - Zabbix module adapter for Go
 * Copyright (C) 2015 - Ryan Armstrong <ryan@cavaliercoder.com>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

// Package main is a shared library with sample Zabbix bindings which may be loaded into a
// Zabbix agent or server using the `LoadModule` directive.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/cavaliercoder/g2z"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// main is a mandatory entry point, although it is never called.
func main() {
	panic("THIS_SHOULD_NEVER_HAPPEN")
}

// init is a mandatory initialization function which must be used exclusively to register Zabbix
// module function handlers. It is called when Zabbix calls `dlopen()` to load this Go module.
//
// No other work should be performed in this function. All other initialization activities should
// be executed with an InitHandlerFunc.
func init() {
	g2z.RegisterInitHandler(InitModule)
	g2z.RegisterUninitHandler(UninitModule)

	g2z.RegisterUint64Item("go.ping", "", Ping)
	g2z.RegisterStringItem("go.echo", "Hello world!", Echo)
	g2z.RegisterDoubleItem("go.random", "0,100", Random)
	g2z.RegisterDiscoveryItem("go.cpu.discovery", "", DiscoverCpus)
}

// InitModule is a InitHandlerFunc which simply add an entry to the Zabbix log.
func InitModule() error {
	g2z.LogInfof("Dummy module initialized")
	return nil
}

// UninitModule is an UninitHandlerFunc which simply add an entry to the Zabbix log.
func UninitModule() error {
	g2z.LogInfof("Dummy module uninitialized")
	return nil
}

// Ping is a Uint64ItemHandlerFunc for key `go.ping` which simply returns 1.
func Ping(request *g2z.AgentRequest) (uint64, error) {
	return 1, nil
}

// Ping is a StringItemHandlerFunc for key `go.echo` which concatenates and returns whatever
// strings are provided as request parameters.
func Echo(request *g2z.AgentRequest) (string, error) {
	return strings.Join(request.Params, " "), nil
}

// Random is a DoubleItemHandlerFunc for key `go.random` which returns a random floating point
// integer within the range of the first and second parameter values.
func Random(request *g2z.AgentRequest) (float64, error) {
	// validate param count
	if len(request.Params) != 2 {
		return 0.00, errors.New("Invalid parameter count")
	}

	// parse first param as float64
	from, err := strconv.ParseFloat(request.Params[0], 64)
	if err != nil {
		return 0.00, err
	}

	// parse second param as float64
	to, err := strconv.ParseFloat(request.Params[1], 64)
	if err != nil {
		return 0.00, err
	}

	// validate range
	if to < from {
		return 0.00, errors.New("Invalid range specified")
	}

	// return a random number in range
	return from + ((to - from) * rand.New(rand.NewSource(time.Now().UnixNano())).Float64()), nil
}

// DiscoveryCpus is a DiscoveryItemHandlerFunc for key `go.cpu.discovery` which returns JSON
// encoding discovery data for all CPUs identified on the host.
func DiscoverCpus(request *g2z.AgentRequest) (g2z.DiscoveryData, error) {
	// init discovery data
	d := make(g2z.DiscoveryData, 0)

	// open /proc/cpuinfo
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// read each line
	i := make(g2z.DiscoveryItem, 0)
	pattern := regexp.MustCompile(`^(.*?)\s*:\s*(.*)$`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()

		// if line is blank, append the last populated item
		if s == "" {
			d = append(d, i)
			i = make(g2z.DiscoveryItem, 0)
		} else if matches := pattern.FindAllStringSubmatch(s, -1); len(matches) > 0 {
			// check if line is a "key    : val" line
			i[matches[0][1]] = matches[0][2]
		} else {
			return nil, errors.New(fmt.Sprintf("Unparsable line in /proc/cpuinfo: \"%s\"", s))
		}
	}

	return d, nil
}