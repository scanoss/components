// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"context"
	"testing"

	common "github.com/scanoss/papi/api/commonv2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetComponentStatistics is deliberately not implemented. What matters is that
// calling it returns an error rather than panicking: the server used to embed the
// papi ComponentsServer interface, whose nil value any unimplemented method
// dereferenced, so one request took the whole process down.
func TestUnimplementedMethodReturnsErrorInsteadOfPanicking(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("calling an unimplemented method panicked: %v", r)
		}
	}()

	server := &componentServer{}
	_, err := server.GetComponentStatistics(context.Background(), &common.ComponentsRequest{})
	if err == nil {
		t.Fatal("want an error from a method the service does not implement")
	}
	if got := status.Code(err); got != codes.Unimplemented {
		t.Errorf("status code = %v, want %v", got, codes.Unimplemented)
	}
}
