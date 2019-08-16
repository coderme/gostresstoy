package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	name    = "GoStressToy"
	version = "0.1"
	license = `
Copyright (c) %d CoderMe.com
 Permission to use, copy, modify, and distribute this software for any
 purpose with or without fee is hereby granted, provided that the above
 copyright notice and this permission notice appear in all copies.
 
 THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
`
)

var (
	concurency  = flag.Int("c", defaultConcurrency, "")
	duration    = flag.Duration("d", 1*time.Minute, "")
	numRequests = flag.Int("n", 0, "")
	countBytes  = flag.Bool("b", false, "")
	dumpJSON    = flag.Bool("j", false, "")
	maxPoolSize = flag.Int("x", maxPool, "")
	help        = flag.Bool("h", false, "")
	showhelp    = flag.Bool("help", false, "")
	vcpu        = flag.Int("v", defaultCPU, "")
	showLicense = flag.Bool("l", false, "")
	showVersion = flag.Bool("version", false, "")
)

func printVersion() {
	fmt.Println(getVersion())
	os.Exit(0)

}

func getVersion() string {
	return name + " v" + version

}

func printLicense() {
	l := "\n" + getVersion() + "\n"
	l += license + "\n"
	fmt.Printf(l, time.Now().Year())
	os.Exit(0)

}

func usage() {
	fmt.Printf(`%s

Usage: %s [-l | -version] [(-d duration | -n total)] [-c concurrent] [-v vcpus] [-j] [-b] URL

FLAGS:
 -version
    Show version and exit.
 -l 
    Show License and exit.
 -h | --help
    Show help and exit.
 -j
    Display stats as JSON (default: false)
 -b
    Count the recieved bytes (default: false)

OPTIONS:
 -d duration
    Do a stress that lasts a duration of time, value is number followed by any single
    character of (s,m,h) which means (second, minute, hour) respectivly, (defaut: 1m) = 1 minute.
 -n total
    Total number of requests to be performed, (if provided it revokes the -d flag, default: 0)
 -c concurrent
    Number of concurrent requests, (default: %d)
 -v vcpus
    Number of vcpus to use during the stress, max of %d, (default: %d)
 -x size
    Maximum number of pregenerated requests, (default: %d)
    Note: number of requests is RAM bound, More requests need More RAM
    for instance: 1000000 (1 million) requests, may reserve upto 1GB RAM or even more, be warned :)

AURGUMENTS:
 URL
    the URL to be stressed

`, getVersion(), os.Args[0], defaultConcurrency, numVCPU, defaultCPU, maxPool)

	os.Exit(1)

}
