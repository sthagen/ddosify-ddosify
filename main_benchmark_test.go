//go:build linux || darwin
// +build linux darwin

/*
*
*	Ddosify - Load testing tool for any web system.
*   Copyright (C) 2021  Ddosify (https://ddosify.com)
*
*   This program is free software: you can redistribute it and/or modify
*   it under the terms of the GNU Affero General Public License as published
*   by the Free Software Foundation, either version 3 of the License, or
*   (at your option) any later version.
*
*   This program is distributed in the hope that it will be useful,
*   but WITHOUT ANY WARRANTY; without even the implied warranty of
*   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*   GNU Affero General Public License for more details.
*
*   You should have received a copy of the GNU Affero General Public License
*   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	gopsProc "github.com/shirou/gopsutil/v3/process"
	"golang.org/x/exp/constraints"
)

type TestType string

const (
	Multipart   TestType = "multipart"
	Correlation TestType = "correlation"
	Basic       TestType = "basic"
)

var table = []struct {
	name             string
	path             string
	cpuTimeThreshold float64
	// in percents
	maxMemThreshold float32
	avgMemThreshold float32

	testType TestType
}{
	{
		name:             "config_distinct_user",
		path:             "config/config_testdata/benchmark/config_distinct_user.json",
		cpuTimeThreshold: 0.350,
		maxMemThreshold:  0.5,
		avgMemThreshold:  0.35,
		testType:         Basic,
	},
	{
		name:             "config_repeated_user",
		path:             "config/config_testdata/benchmark/config_repeated_user.json",
		cpuTimeThreshold: 0.350,
		maxMemThreshold:  0.5,
		avgMemThreshold:  0.35,
		testType:         Basic,
	},
	{
		name:             "config_correlation_load_1",
		path:             "config/config_testdata/benchmark/config_correlation_load_1.json",
		cpuTimeThreshold: 0.350,
		maxMemThreshold:  0.5,
		avgMemThreshold:  0.35,
		testType:         Correlation,
	},
	{
		name:             "config_correlation_load_2",
		path:             "config/config_testdata/benchmark/config_correlation_load_2.json",
		cpuTimeThreshold: 2.5,
		maxMemThreshold:  0.8,
		avgMemThreshold:  0.7,
		testType:         Correlation,
	},
	{
		name:             "config_correlation_load_3",
		path:             "config/config_testdata/benchmark/config_correlation_load_3.json",
		cpuTimeThreshold: 15.5,
		maxMemThreshold:  5,
		avgMemThreshold:  4,
		testType:         Correlation,
	},
	{
		name:             "config_correlation_load_4",
		path:             "config/config_testdata/benchmark/config_correlation_load_4.json",
		cpuTimeThreshold: 25,
		maxMemThreshold:  7,
		avgMemThreshold:  5,
		testType:         Correlation,
	},
	{
		name:             "config_correlation_load_5",
		path:             "config/config_testdata/benchmark/config_correlation_load_5.json",
		cpuTimeThreshold: 60,
		maxMemThreshold:  15,
		avgMemThreshold:  10,
		testType:         Correlation,
	},
	//{
	//	name:             "config_multipart_inject_10rps",
	//	path:             "config/config_testdata/benchmark/config_multipart_inject_10rps.json",
	//	cpuTimeThreshold: 5,
	//	maxMemThreshold:  2,
	//	avgMemThreshold:  1,
	//	testType:         Multipart,
	//},
	//{
	//	name:             "config_multipart_inject_100rps",
	//	path:             "config/config_testdata/benchmark/config_multipart_inject_100rps.json",
	//	cpuTimeThreshold: 50,
	//	maxMemThreshold:  3,
	//	avgMemThreshold:  2,
	//	testType:         Multipart,
	//},
	//{
	//	name:             "config_multipart_inject_200rps",
	//	path:             "config/config_testdata/benchmark/config_multipart_inject_200rps.json",
	//	cpuTimeThreshold: 100,
	//	maxMemThreshold:  5,
	//	avgMemThreshold:  4,
	//	testType:         Multipart,
	//},
	//{
	//	name:             "config_multipart_inject_500rps",
	//	path:             "config/config_testdata/benchmark/config_multipart_inject_500rps.json",
	//	cpuTimeThreshold: 160,
	//	maxMemThreshold:  7,
	//	avgMemThreshold:  10,
	//	testType:         Multipart,
	//},
	{
		name:             "config_multipart_inject_1krps",
		path:             "config/config_testdata/benchmark/config_multipart_inject_1krps.json",
		cpuTimeThreshold: 200,
		maxMemThreshold:  10,
		avgMemThreshold:  15,
		testType:         Multipart,
	},
	//{
	//	name:             "config_multipart_inject_2krps",
	//	path:             "config/config_testdata/benchmark/config_multipart_inject_2krps.json",
	//	cpuTimeThreshold: 300,
	//	maxMemThreshold:  15,
	//	avgMemThreshold:  20,
	//	testType: Multipart,
	//},
}

var cpuprofile = flag.String("cpuprof", "", "write cpu profiles")
var memprofile = flag.String("memprof", "", "write memory profiles")
var keepTrace = flag.String("tracef", "", "write execution traces")
var runBenchmarkN = flag.Int("runN", 1, "run benchmarks N times")

func BenchmarkEngines(t *testing.B) {
	index := os.Getenv("index")
	if index == "" {
		N := 1
		if *runBenchmarkN > 1 {
			N = *runBenchmarkN
		}
		// parent
		success := true
		originalN := N
		for i, _ := range table { // open a new process for each test config
			N = originalN
			if table[i].testType != Multipart {
				N = 1 // if not multipart, run only once
			}
			for j := 0; j < N; j++ { // run each test config N times
				time.Sleep(1 * time.Second) // wait for the previous process to finish
				// start a child
				env := fmt.Sprintf("index=%d", i)
				cPid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{Files: []uintptr{0, 1, 2}, Env: []string{env}})
				if err != nil {
					panic(err.Error())
				}

				proc, err := os.FindProcess(cPid)
				if err != nil {
					panic(err.Error())
				}

				pState, err := proc.Wait()
				if err != nil {
					panic(err.Error())
				}

				if !pState.Success() {
					success = false
				}
			}
			if !success {
				t.Fail()
			}
		}

	} else {
		i, _ := strconv.Atoi(index)
		conf := table[i]
		outSuffix := ".out"
		var err error

		// child proc
		var cpuProfFile, memProfFile, traceFile *os.File
		if *cpuprofile != "" {
			cpuProfFile, err = os.Create(fmt.Sprintf("%s_cpuprof_%s.out", strings.TrimSuffix(*cpuprofile, outSuffix), conf.name))
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(cpuProfFile)
			defer cpuProfFile.Close()
			defer pprof.StopCPUProfile()
		}

		if *memprofile != "" { // get memory profile at execution finish
			memProfFile, err = os.Create(fmt.Sprintf("%s_memprof_%s.out", strings.TrimSuffix(*memprofile, outSuffix),
				conf.name))
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer memProfFile.Close() // error handling omitted for example
			defer func() {
				pprof.Lookup("allocs").WriteTo(memProfFile, 0)
				// if you want to check live heap objects:
				// runtime.GC() // get up-to-date statistics
				// pprof.Lookup("heap").WriteTo(memProfFile, 0)
			}()
		}

		if *keepTrace != "" {
			traceFile, err = os.Create(fmt.Sprintf("%s_trace_%s.out", strings.TrimSuffix(*keepTrace, outSuffix), conf.name))
			if err != nil {
				log.Fatalf("failed to create trace output file: %v", err)
			}
			defer func() {
				if err := traceFile.Close(); err != nil {
					log.Fatalf("failed to close trace file: %v", err)
				}
			}()

			if err := trace.Start(traceFile); err != nil {
				log.Fatalf("failed to start trace: %v", err)
			}
			defer trace.Stop()
		}

		success := t.Run(fmt.Sprintf("config_%s", conf.path), func(t *testing.B) {
			var memPercents []float32
			var cpuStats []*cpu.TimesStat

			*configPath = conf.path
			run = tempRun
			doneChan := make(chan struct{}, 1)
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				pid := os.Getpid()
				proc, _ := gopsProc.NewProcess(int32(pid))
				for {
					select {
					case <-ticker.C:
						cpuStat, _ := proc.Times()
						cpuStats = append(cpuStats, cpuStat)

						memPerc, _ := proc.MemoryPercent()
						memPercents = append(memPercents, memPerc)
					case <-doneChan:
						return
					}
				}
			}()
			start()
			doneChan <- struct{}{}

			lastCpuStat := cpuStats[len(cpuStats)-1]
			cpuTime := lastCpuStat.User + lastCpuStat.System
			fmt.Printf("cpuTime: %f / %f \n", cpuTime, conf.cpuTimeThreshold)

			avgMem := sum(memPercents) / float32(len(memPercents))
			maxMem := max(memPercents)
			fmt.Printf("Avg mem: %f / %f \n", avgMem, conf.avgMemThreshold)
			fmt.Printf("Max mem: %f / %f \n\n", maxMem, conf.maxMemThreshold)

			if cpuTime > conf.cpuTimeThreshold {
				t.Errorf("Cpu time %f, higher than cpuTimeThreshold %f", cpuTime, conf.cpuTimeThreshold)
			}
			if avgMem > conf.avgMemThreshold {
				t.Errorf("Avg mem %f, higher than avgMemThreshold %f", avgMem, conf.avgMemThreshold)
			}
			if maxMem > conf.maxMemThreshold {
				t.Errorf("Max mem %f, higher than maxMemThreshold %f", maxMem, conf.maxMemThreshold)
			}

		})

		if !success {
			runtime.Goexit()
		}
	}
}

func max[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	m := s[0]
	for _, v := range s {
		if m < v {
			m = v
		}
	}
	return m
}

func sum[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	var m T
	for _, v := range s {
		m += v
	}
	return m
}
