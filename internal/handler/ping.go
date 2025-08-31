// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
