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

package report

import (
	"fmt"
	"reflect"

	"go.ddosify.com/ddosify/core/assertion"
	"go.ddosify.com/ddosify/core/types"
)

var AvailableOutputServices = make(map[string]ReportService)

// ReportService is the interface that abstracts different report implementations.
type ReportService interface {
	DoneChan() <-chan bool
	Init(debug bool, samplingRate int) error
	Start(input chan *types.ScenarioResult, assertionResultChan <-chan assertion.TestAssertionResult)
}

// NewReportService is the factory method of the ReportService.
func NewReportService(s string) (service ReportService, err error) {
	if val, ok := AvailableOutputServices[s]; ok {
		// Create a new object from the service type
		service = reflect.New(reflect.TypeOf(val).Elem()).Interface().(ReportService)
	} else {
		err = fmt.Errorf("unsupported output type: %s", s)
	}

	return
}
