/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"bytes"
	"io"
	"testing"
	"time"
)

func TestTimeoutEOF(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0, 1, 2, 3})
	trw := NewTimeoutReadWriter(buf, 10*time.Second, 0)

	out := make([]byte, 10)
	n, err := trw.Read(out)
	if n != 4 {
		t.Errorf("Read %d bytes; want 4 bytes", n)
	}

	if err != io.EOF {
		t.Errorf("Error was %v; want EOF", err)
	}
}
