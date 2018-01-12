# GoStressToy

## About
* a command line "toy" program which can bench/stress http/https webservers, its cpu/ram bound

## Features
* Can generate load using fixed number of requests.
* Or maintain stress for a certain duration of time.
* Configurable concurency level, vcpu utilized
* Option to display stats as JSON.

## Usage
* clone or download the program from "bin" thats suitable for your os/arch
* from command line run the program with -h for help

## Examples
* Stressing localhost using 2 vcpus, concurrency 50 for duration of 3 minutes
 gostressstoy -d 3m -v 2 -c 50 http://localhost

## DISCLAIMER
* Benchmarking and/or Stressing other's people websites without their permission is unethical and criminal act.

## LICENSE
* Check the -l ( lower L) flag of the program
* or the LICENSE file.






