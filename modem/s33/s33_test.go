// Copyright 2020 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s33

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/wathiede/surfer/modem"
)

func TestParseStatus(t *testing.T) {
	flag.Set("v", "true")
	flag.Set("logtostderr", "true")

	p := "testdata/S33-signal.json"
	r, err := os.Open(p)
	if err != nil {
		t.Fatalf("Failed to open %q: %v", p, err)
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read test file")
	}
	status := statusResponse{}
	err = json.Unmarshal(data, &status)
	if err != nil {
		t.Fatalf("Unable to parse JSON")
	}
	got, err := parseStatus(&status)
	if err != nil {
		t.Fatalf("Failed to parse %q: %v", p, err)
	}

	want := &modem.Signal{
		Downstream: map[modem.Channel]*modem.Downstream{
			"1": {
				Modulation:    "QAM256",
				Frequency:     "441000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"2": {
				Modulation:    "QAM256",
				Frequency:     "447000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"3": {
				Modulation:    "QAM256",
				Frequency:     "453000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"4": {
				Modulation:    "QAM256",
				Frequency:     "459000000 Hz",
				PowerLevel:    -4,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"5": {
				Modulation:    "QAM256",
				Frequency:     "465000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"6": {
				Modulation:    "QAM256",
				Frequency:     "471000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"7": {
				Modulation:    "QAM256",
				Frequency:     "477000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"8": {
				Modulation:    "QAM256",
				Frequency:     "483000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"9": {
				Modulation:    "QAM256",
				Frequency:     "489000000 Hz",
				PowerLevel:    -3,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"10": {
				Modulation:    "QAM256",
				Frequency:     "507000000 Hz",
				PowerLevel:    -4,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"11": {
				Modulation:    "QAM256",
				Frequency:     "513000000 Hz",
				PowerLevel:    -4,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"12": {
				Modulation:    "QAM256",
				Frequency:     "519000000 Hz",
				PowerLevel:    -4,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"13": {
				Modulation:    "QAM256",
				Frequency:     "525000000 Hz",
				PowerLevel:    -4,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"14": {
				Modulation:    "QAM256",
				Frequency:     "531000000 Hz",
				PowerLevel:    -4,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"15": {
				Modulation:    "QAM256",
				Frequency:     "537000000 Hz",
				PowerLevel:    -4,
				SNR:           40,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"16": {
				Modulation:    "QAM256",
				Frequency:     "543000000 Hz",
				PowerLevel:    -4,
				SNR:           38,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"17": {
				Modulation:    "QAM256",
				Frequency:     "549000000 Hz",
				PowerLevel:    -4,
				SNR:           40,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"18": {
				Modulation:    "QAM256",
				Frequency:     "555000000 Hz",
				PowerLevel:    -4,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"19": {
				Modulation:    "QAM256",
				Frequency:     "561000000 Hz",
				PowerLevel:    -4,
				SNR:           43,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"20": {
				Modulation:    "QAM256",
				Frequency:     "567000000 Hz",
				PowerLevel:    -4,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"21": {
				Modulation:    "QAM256",
				Frequency:     "573000000 Hz",
				PowerLevel:    -4,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"22": {
				Modulation:    "QAM256",
				Frequency:     "579000000 Hz",
				PowerLevel:    -5,
				SNR:           41,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"23": {
				Modulation:    "QAM256",
				Frequency:     "585000000 Hz",
				PowerLevel:    -5,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"24": {
				Modulation:    "QAM256",
				Frequency:     "591000000 Hz",
				PowerLevel:    -5,
				SNR:           41,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"25": {
				Modulation:    "OFDM PLC",
				Frequency:     "693000000 Hz",
				PowerLevel:    -4,
				SNR:           41,
				Correctable:   590747125,
				Uncorrectable: 0,
			},
			"26": {
				Modulation:    "QAM256",
				Frequency:     "597000000 Hz",
				PowerLevel:    -5,
				SNR:           38,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"27": {
				Modulation:    "QAM256",
				Frequency:     "603000000 Hz",
				PowerLevel:    -5,
				SNR:           40,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"28": {
				Modulation:    "QAM256",
				Frequency:     "609000000 Hz",
				PowerLevel:    -5,
				SNR:           41,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"29": {
				Modulation:    "QAM256",
				Frequency:     "615000000 Hz",
				PowerLevel:    -5,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"30": {
				Modulation:    "QAM256",
				Frequency:     "621000000 Hz",
				PowerLevel:    -5,
				SNR:           41,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"31": {
				Modulation:    "QAM256",
				Frequency:     "627000000 Hz",
				PowerLevel:    -5,
				SNR:           41,
				Correctable:   0,
				Uncorrectable: 0,
			},
			"32": {
				Modulation:    "QAM256",
				Frequency:     "633000000 Hz",
				PowerLevel:    -5,
				SNR:           42,
				Correctable:   0,
				Uncorrectable: 0,
			},
		},
		Upstream: map[modem.Channel]*modem.Upstream{
			"5": {
				Frequency:  "36500000 Hz",
				PowerLevel: 46.8,
				Modulation: "SC-QAM",
				Status:     "Locked",
			},
			"6": {
				Frequency:  "30100000 Hz",
				PowerLevel: 46.3,
				Modulation: "SC-QAM",
				Status:     "Not Locked",
			},
			"7": {
				Frequency:  "23700000 Hz",
				PowerLevel: 44.0,
				Modulation: "SC-QAM",
				Status:     "Not Locked",
			},
			"8": {
				Frequency:  "17300000 Hz",
				PowerLevel: 41.8,
				Modulation: "SC-QAM",
				Status:     "Not Locked",
			},
		},
	}

	if !reflect.DeepEqual(want, got) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("Got:\n%s\nWant:\n%s", g, w)
	}
}
