package encoding

import (
	"bufio"
	"bytes"
	"strconv"

	"github.com/pkg/errors"
)

func syslogParser() bufio.SplitFunc {
	// format:
	//nolint:lll
	// 64 <190>1 2019-07-20T17:50:10.879238Z shuttle token shuttle - - 99\n65 <190>1 2019-07-20T17:50:10.879238Z shuttle token shuttle - - 100\n
	// ^ frame size                                                       ^ boundary
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// first space gives us the frame size
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			if atEOF && len(data) > 0 {
				return 0, nil, errors.Wrap(ErrBadFrame, "missing frame length")
			}
			return 0, nil, nil
		}

		if sp == 0 {
			return 0, nil, errors.Wrap(ErrBadFrame, "invalid frame length")
		}

		msgSize, err := strconv.ParseUint(string(data[0:sp]), 10, 64)
		if err != nil {
			return 0, nil, errors.Wrap(ErrBadFrame, "couldnt parse frame length")
		}

		// 1 here is the 'space' itself, used in the framing above
		dataBoundary := sp + int(msgSize) + 1

		if dataBoundary > len(data) {
			if atEOF {
				return 0, nil, errors.Wrap(ErrBadFrame, "message boundary not respected")
			}
			return 0, nil, nil
		}

		return dataBoundary, data[sp+1 : dataBoundary], nil
	}
}
